package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/hatolife/HatoNote/internal/document"
	"github.com/hatolife/HatoNote/internal/markdownquality"
	"github.com/hatolife/HatoNote/internal/settings"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type DocumentDiagnostic struct {
	Type    string `json:"type"`
	Page    string `json:"page"`
	Line    int    `json:"line,omitempty"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message"`
}

type documentReference struct {
	target string
	image  bool
}

func markdownReferences(source []byte) []documentReference {
	reader := text.NewReader(source)
	root := goldmark.New().Parser().Parse(reader)
	result := []documentReference{}
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := node.(type) {
		case *ast.Link:
			result = append(result, documentReference{target: string(n.Destination)})
		case *ast.Image:
			result = append(result, documentReference{target: string(n.Destination), image: true})
		}
		return ast.WalkContinue, nil
	})
	return result
}

func relativeReferenceTarget(raw string) (string, bool) {
	target := strings.TrimSpace(raw)
	if target == "" || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") {
		return "", false
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" {
		return "", false
	}
	target = parsed.Path
	if target == "" || strings.HasPrefix(target, "/") || strings.Contains(target, "\\") {
		return "", false
	}
	return target, true
}

func bundleReferenceTarget(pageName, target string) (string, bool) {
	resolved := path.Clean(path.Join(path.Dir(pageName), target))
	if resolved == ".." || strings.HasPrefix(resolved, "../") || path.IsAbs(resolved) {
		return "", false
	}
	return resolved, true
}

func imageAsset(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg":
		return true
	default:
		return false
	}
}

func localReferenceExists(filename, target string) bool {
	if strings.TrimSpace(filename) == "" {
		return false
	}
	_, err := os.Stat(filepath.Clean(filepath.Join(filepath.Dir(filename), filepath.FromSlash(target))))
	return err == nil
}

func diagnoseDocument(doc *bundle.Document, filename string, documentType document.Type, cfg settings.Settings) []DocumentDiagnostic {
	if doc == nil {
		return []DocumentDiagnostic{}
	}
	result := []DocumentDiagnostic{}
	usedImages := map[string]bool{}
	for _, pageName := range doc.Pages() {
		source, ok := doc.Files[pageName]
		if !ok || !bundle.IsMarkdown(pageName) {
			continue
		}
		for _, diagnostic := range markdownquality.Lint(string(source), markdownLintOptions(cfg)) {
			result = append(result, DocumentDiagnostic{
				Type: "lint", Page: pageName, Line: diagnostic.Line, Rule: diagnostic.Rule, Message: diagnostic.Message,
			})
		}
		for _, reference := range markdownReferences(source) {
			target, relative := relativeReferenceTarget(reference.target)
			if !relative {
				continue
			}
			resolved := target
			missing := false
			if documentType == document.Markdown {
				missing = !localReferenceExists(filename, target)
			} else {
				var inside bool
				resolved, inside = bundleReferenceTarget(pageName, target)
				if !inside {
					missing = true
				} else {
					_, exists := doc.Files[resolved]
					missing = !exists
				}
			}
			if missing {
				if reference.image {
					result = append(result, DocumentDiagnostic{Type: "broken-image", Page: pageName, Message: "画像が見つかりません: " + target})
				} else {
					result = append(result, DocumentDiagnostic{Type: "broken-link", Page: pageName, Message: "リンク先が見つかりません: " + target})
				}
				continue
			}
			if reference.image && documentType != document.Markdown && imageAsset(resolved) {
				usedImages[resolved] = true
			}
		}
	}
	if documentType != document.Markdown {
		images := []string{}
		for name := range doc.Files {
			if imageAsset(name) {
				images = append(images, name)
			}
		}
		sort.Strings(images)
		for _, name := range images {
			if !usedImages[name] {
				result = append(result, DocumentDiagnostic{Type: "unused-image", Message: "未使用画像: " + name})
			}
		}
	}
	return result
}

// DocumentDiagnostics は現在の未保存内容を取り込んで文書全体を診断します。
func (a *App) DocumentDiagnostics() ([]DocumentDiagnostic, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session == nil {
		return nil, fmt.Errorf("文書を開いてください")
	}
	if err := a.syncNativeLocked(); err != nil {
		return nil, err
	}
	return diagnoseDocument(a.session.Doc, a.session.Filename, a.documentTypeLocked(), a.cfg), nil
}
