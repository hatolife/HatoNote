package main

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/hatolife/HatoNote/internal/document"
)

// MovePage は並べ替え前の順序を確認し、ファイル名を変えずにページを移動します。
func (a *App) MovePage(expected []string, name, target string, after bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session == nil || a.documentTypeLocked() != document.MDZ {
		return fmt.Errorf("通常MDZを開いてください")
	}
	if err := a.syncNativeLocked(); err != nil {
		return err
	}
	if err := a.session.Capture(); err != nil {
		return err
	}
	order := a.session.Doc.Pages()
	if !slices.Equal(order, expected) {
		return fmt.Errorf("ページ構成が変更されました。最新の一覧を確認してください")
	}
	from, to := slices.Index(order, name), slices.Index(order, target)
	if from < 0 || to < 0 {
		return fmt.Errorf("移動対象のページがありません")
	}
	if from == to {
		return nil
	}
	order = slices.Delete(order, from, from+1)
	to = slices.Index(order, target)
	if after {
		to++
	}
	order = slices.Insert(order, to, name)
	return a.session.SetPageOrder(order)
}


// DuplicatePage は通常MDZのMarkdownページを複製し、元ページの直後へ挿入します。
func (a *App) DuplicatePage(name string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session == nil || a.documentTypeLocked() != document.MDZ {
		return "", fmt.Errorf("通常MDZを開いてください")
	}
	if err := a.syncNativeLocked(); err != nil {
		return "", err
	}
	if err := a.session.Capture(); err != nil {
		return "", err
	}
	data, ok := a.session.Doc.Files[name]
	if !ok || !strings.EqualFold(path.Ext(name), ".md") && !strings.EqualFold(path.Ext(name), ".markdown") {
		return "", fmt.Errorf("複製するMarkdownページがありません")
	}
	order := a.session.Doc.Pages()
	index := slices.Index(order, name)
	if index < 0 {
		return "", fmt.Errorf("複製するページが一覧にありません")
	}
	dir, ext := path.Dir(name), path.Ext(name)
	if dir == "." {
		dir = ""
	}
	stem := strings.TrimSuffix(path.Base(name), ext)
	used := map[string]bool{}
	for existing := range a.session.Doc.Files {
		used[strings.ToLower(existing)] = true
	}
	candidate := ""
	for n := 1; ; n++ {
		suffix := "-copy"
		if n > 1 {
			suffix = fmt.Sprintf("-copy-%d", n)
		}
		candidate = path.Join(dir, stem+suffix+ext)
		if !used[strings.ToLower(candidate)] {
			break
		}
	}
	if err := a.session.Put(candidate, append([]byte(nil), data...)); err != nil {
		return "", err
	}
	order = slices.Insert(order, index+1, candidate)
	if err := a.session.SetPageOrder(order); err != nil {
		return "", err
	}
	return candidate, nil
}
