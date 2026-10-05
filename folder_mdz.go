package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// MarkdownFolderInfo はフォルダからMDZを生成できるか判断するための情報です。
type MarkdownFolderInfo struct {
	Path          string `json:"path"`
	Name          string `json:"name"`
	MarkdownCount int    `json:"markdownCount"`
	FileCount     int    `json:"fileCount"`
	Directory     bool   `json:"directory"`
}

// MarkdownFolderInfo は指定パスがフォルダなら配下のMarkdown件数を調べます。
func (a *App) MarkdownFolderInfo(folder string) (MarkdownFolderInfo, error) {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		return MarkdownFolderInfo{}, nil
	}
	absolute, err := filepath.Abs(folder)
	if err != nil {
		return MarkdownFolderInfo{}, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return MarkdownFolderInfo{}, err
	}
	if !info.IsDir() {
		return MarkdownFolderInfo{Path: absolute, Name: info.Name()}, nil
	}
	return inspectMarkdownFolder(absolute)
}

// CreateMDZFromFolder はフォルダ配下をMDZへまとめ、選択した保存先へ書き出します。
func (a *App) CreateMDZFromFolder(folder string) (string, error) {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		var err error
		folder, err = runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "MDZにするフォルダを選択"})
		if err != nil {
			return "", err
		}
		if folder == "" {
			return "", nil
		}
	}
	doc, info, err := buildMDZDocumentFromFolder(folder)
	if err != nil {
		return "", err
	}
	defaultName := info.Name + ".mdz"
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: "フォルダからMDZを生成", DefaultFilename: defaultName, Filters: []runtime.FileFilter{{DisplayName: "MDZip", Pattern: "*.mdz"}}})
	if err != nil {
		return "", err
	}
	if filename == "" {
		return "", nil
	}
	if !strings.EqualFold(filepath.Ext(filename), ".mdz") {
		filename += ".mdz"
	}
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	if err := doc.Write(absolute, version); err != nil {
		return "", err
	}
	return absolute, nil
}

func inspectMarkdownFolder(folder string) (MarkdownFolderInfo, error) {
	absolute, err := filepath.Abs(folder)
	if err != nil {
		return MarkdownFolderInfo{}, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return MarkdownFolderInfo{}, err
	}
	if !info.IsDir() {
		return MarkdownFolderInfo{Path: absolute, Name: info.Name()}, nil
	}
	result := MarkdownFolderInfo{Path: absolute, Name: info.Name(), Directory: true}
	err = walkMarkdownFolder(absolute, func(name, source string) error {
		result.FileCount++
		if bundle.IsMarkdown(name) {
			result.MarkdownCount++
		}
		return nil
	})
	return result, err
}

func buildMDZDocumentFromFolder(folder string) (*bundle.Document, MarkdownFolderInfo, error) {
	info, err := inspectMarkdownFolder(folder)
	if err != nil {
		return nil, MarkdownFolderInfo{}, err
	}
	if !info.Directory {
		return nil, info, fmt.Errorf("フォルダを指定してください")
	}
	if info.MarkdownCount == 0 {
		return nil, info, fmt.Errorf("フォルダ配下にMarkdownがありません")
	}
	doc := bundle.New()
	markdowns := make([]string, 0, info.MarkdownCount)
	err = walkMarkdownFolder(info.Path, func(name, source string) error {
		data, err := readLimited(source)
		if err != nil {
			return err
		}
		if err := doc.Put(name, data); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if bundle.IsMarkdown(name) {
			markdowns = append(markdowns, name)
		}
		return nil
	})
	if err != nil {
		return nil, info, err
	}
	sort.SliceStable(markdowns, func(i, j int) bool {
		if naturalFilenameLess(markdowns[i], markdowns[j]) {
			return true
		}
		if naturalFilenameLess(markdowns[j], markdowns[i]) {
			return false
		}
		return strings.ToLower(markdowns[i]) < strings.ToLower(markdowns[j])
	})
	doc.Entry = preferredFolderEntry(markdowns)
	order, err := json.Marshal(markdowns)
	if err != nil {
		return nil, info, err
	}
	doc.Manifest[bundle.PageOrderKey] = order
	return doc, info, nil
}

func walkMarkdownFolder(root string, visit func(name, source string) error) error {
	return filepath.Walk(root, func(current string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if current == root {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			switch strings.ToLower(info.Name()) {
			case ".git", ".hg", ".svn":
				return filepath.SkipDir
			default:
				return nil
			}
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if strings.EqualFold(name, "manifest.json") {
			return nil
		}
		if !bundle.ValidPath(name) {
			return fmt.Errorf("MDZへ格納できないパスです: %s", name)
		}
		return visit(name, current)
	})
}

func preferredFolderEntry(markdowns []string) string {
	for _, preferred := range []string{"README.md", "README.markdown", "index.md", "index.markdown"} {
		for _, name := range markdowns {
			if !strings.Contains(name, "/") && strings.EqualFold(name, preferred) {
				return name
			}
		}
	}
	if len(markdowns) == 0 {
		return ""
	}
	return markdowns[0]
}
