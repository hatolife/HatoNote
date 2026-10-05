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

// MarkdownFolderInfo はフォルダや複数選択からMDZを生成できるか判断するための情報です。
type MarkdownFolderInfo struct {
	Path           string `json:"path"`
	Name           string `json:"name"`
	MarkdownCount  int    `json:"markdownCount"`
	FileCount      int    `json:"fileCount"`
	Directory      bool   `json:"directory"`
	SelectionCount int    `json:"selectionCount"`
}

type mdzSource struct {
	path  string
	isDir bool
}

type mdzSourceSet struct {
	base    string
	sources []mdzSource
	info    MarkdownFolderInfo
}

// MarkdownFolderInfo は指定パスがフォルダなら配下のMarkdown件数を調べます。
func (a *App) MarkdownFolderInfo(folder string) (MarkdownFolderInfo, error) {
	return a.MarkdownPathsInfo([]string{folder})
}

// MarkdownPathsInfo は複数のファイルとフォルダを1つのMDZ生成対象として調べます。
func (a *App) MarkdownPathsInfo(paths []string) (MarkdownFolderInfo, error) {
	set, err := prepareMDZSources(paths)
	if err != nil {
		return MarkdownFolderInfo{}, err
	}
	if len(set.sources) == 0 {
		return MarkdownFolderInfo{}, nil
	}
	return set.info, nil
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
	return a.createMDZFromPaths([]string{folder})
}

// CreateMDZFromPaths は複数選択の階層構造を維持して1つのMDZへまとめます。
func (a *App) CreateMDZFromPaths(paths []string) (string, error) {
	return a.createMDZFromPaths(paths)
}

func (a *App) createMDZFromPaths(paths []string) (string, error) {
	doc, info, err := buildMDZDocumentFromPaths(paths)
	if err != nil {
		return "", err
	}
	defaultName := info.Name + ".mdz"
	if info.Name == "" {
		defaultName = "document.mdz"
	}
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
	set, err := prepareMDZSources([]string{folder})
	if err != nil {
		return MarkdownFolderInfo{}, err
	}
	return set.info, nil
}

func buildMDZDocumentFromFolder(folder string) (*bundle.Document, MarkdownFolderInfo, error) {
	return buildMDZDocumentFromPaths([]string{folder})
}

func buildMDZDocumentFromPaths(paths []string) (*bundle.Document, MarkdownFolderInfo, error) {
	set, err := prepareMDZSources(paths)
	if err != nil {
		return nil, MarkdownFolderInfo{}, err
	}
	if len(set.sources) == 0 {
		return nil, set.info, fmt.Errorf("MDZにするファイルまたはフォルダを指定してください")
	}
	if !set.info.Directory {
		return nil, set.info, fmt.Errorf("フォルダを含む選択を指定してください")
	}
	if set.info.MarkdownCount == 0 {
		return nil, set.info, fmt.Errorf("選択項目の配下にMarkdownがありません")
	}
	doc := bundle.New()
	markdowns := make([]string, 0, set.info.MarkdownCount)
	err = walkMDZSources(set, func(name, source string) error {
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
		return nil, set.info, err
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
		return nil, set.info, err
	}
	doc.Manifest[bundle.PageOrderKey] = order
	return doc, set.info, nil
}

func prepareMDZSources(paths []string) (mdzSourceSet, error) {
	sources, err := normalizeMDZSources(paths)
	if err != nil {
		return mdzSourceSet{}, err
	}
	if len(sources) == 0 {
		return mdzSourceSet{}, nil
	}
	base, err := mdzSourceBase(sources)
	if err != nil {
		return mdzSourceSet{}, err
	}
	info := MarkdownFolderInfo{
		Path:           base,
		Name:           filepath.Base(base),
		SelectionCount: len(sources),
	}
	for _, source := range sources {
		if source.isDir {
			info.Directory = true
			break
		}
	}
	if len(sources) == 1 && sources[0].isDir {
		info.Path = sources[0].path
		info.Name = filepath.Base(sources[0].path)
	}
	set := mdzSourceSet{base: base, sources: sources, info: info}
	err = walkMDZSources(set, func(name, source string) error {
		set.info.FileCount++
		if bundle.IsMarkdown(name) {
			set.info.MarkdownCount++
		}
		return nil
	})
	return set, err
}

func normalizeMDZSources(paths []string) ([]mdzSource, error) {
	items := make([]mdzSource, 0, len(paths))
	seen := map[string]bool{}
	for _, value := range paths {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		absolute, err := filepath.Abs(value)
		if err != nil {
			return nil, err
		}
		absolute = filepath.Clean(absolute)
		if seen[absolute] {
			continue
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			continue
		}
		seen[absolute] = true
		items = append(items, mdzSource{path: absolute, isDir: info.IsDir()})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if len(items[i].path) != len(items[j].path) {
			return len(items[i].path) < len(items[j].path)
		}
		return items[i].path < items[j].path
	})
	result := make([]mdzSource, 0, len(items))
	for _, item := range items {
		covered := false
		for _, parent := range result {
			if parent.isDir && pathWithin(parent.path, item.path) {
				covered = true
				break
			}
		}
		if !covered {
			result = append(result, item)
		}
	}
	return result, nil
}

func mdzSourceBase(sources []mdzSource) (string, error) {
	if len(sources) == 1 && sources[0].isDir {
		return sources[0].path, nil
	}
	parents := make([]string, 0, len(sources))
	for _, source := range sources {
		parents = append(parents, filepath.Dir(source.path))
	}
	base := filepath.Clean(parents[0])
	for _, candidate := range parents[1:] {
		for !pathWithin(base, candidate) {
			parent := filepath.Dir(base)
			if parent == base {
				return "", fmt.Errorf("選択項目に共通の親フォルダがありません")
			}
			base = parent
		}
	}
	return base, nil
}

func pathWithin(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil || filepath.IsAbs(relative) {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)))
}

func walkMDZSources(set mdzSourceSet, visit func(name, source string) error) error {
	for _, source := range set.sources {
		if source.isDir {
			if isIgnoredMDZDirectory(filepath.Base(source.path)) {
				continue
			}
			err := filepath.Walk(source.path, func(current string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.Mode()&os.ModeSymlink != 0 {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				if info.IsDir() {
					if current != source.path && isIgnoredMDZDirectory(info.Name()) {
						return filepath.SkipDir
					}
					return nil
				}
				if !info.Mode().IsRegular() {
					return nil
				}
				return visitMDZSourceFile(set.base, current, visit)
			})
			if err != nil {
				return err
			}
			continue
		}
		if err := visitMDZSourceFile(set.base, source.path, visit); err != nil {
			return err
		}
	}
	return nil
}

func visitMDZSourceFile(base, source string, visit func(name, source string) error) error {
	relative, err := filepath.Rel(base, source)
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
	return visit(name, source)
}

func isIgnoredMDZDirectory(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".hg", ".svn":
		return true
	default:
		return false
	}
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
