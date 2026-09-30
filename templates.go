package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/hatolife/HatoNote/internal/document"
	"github.com/hatolife/HatoNote/internal/workspace"
)

type PageTemplate struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

var builtInPageTemplates = []PageTemplate{
	{ID: "blank", Title: "空ページ", Content: "# 新しいページ\n"},
	{ID: "memo", Title: "メモ", Content: "# メモ\n\n## 内容\n\n"},
	{ID: "meeting", Title: "議事録", Content: "# 議事録\n\n- 日時:\n- 参加者:\n\n## 議題\n\n## 決定事項\n\n## TODO\n\n"},
	{ID: "spec", Title: "仕様書", Content: "# 仕様書\n\n## 目的\n\n## 要件\n\n## 仕様\n\n## 備考\n\n"},
}

func (a *App) PageTemplates() []PageTemplate {
	result := make([]PageTemplate, len(builtInPageTemplates))
	for i, template := range builtInPageTemplates {
		result[i] = PageTemplate{ID: template.ID, Title: template.Title}
	}
	return result
}

func pageTemplateContent(id string) (string, error) {
	for _, template := range builtInPageTemplates {
		if template.ID == id {
			return template.Content, nil
		}
	}
	return "", fmt.Errorf("テンプレートが見つかりません")
}

func (a *App) addPageLocked(name, text string) error {
	if a.session == nil {
		return fmt.Errorf("文書を開いてください")
	}
	documentType := a.documentTypeLocked()
	if documentType == document.Slides {
		return fmt.Errorf("スライドを追加してください")
	}
	if !documentType.Capabilities().MultiplePages {
		return fmt.Errorf("この文書種別ではページを追加できません")
	}
	name = strings.TrimSpace(name)
	if name != "" && path.Ext(name) == "" && !strings.HasSuffix(name, "/") {
		name += ".md"
	}
	if !bundle.IsMarkdown(name) {
		return fmt.Errorf("拡張子を.mdにしてください")
	}
	if err := a.syncNativeLocked(); err != nil {
		return err
	}
	if err := a.session.Capture(); err != nil {
		return err
	}
	for existing := range a.session.Doc.Files {
		if strings.EqualFold(existing, name) {
			return fmt.Errorf("同名のファイルがあります")
		}
	}
	order := a.session.Doc.Pages()
	if err := a.session.Put(name, []byte(text)); err != nil {
		return err
	}
	return a.session.SetPageOrder(append(order, name))
}

func (a *App) AddPageFromTemplate(name, templateID string) error {
	text, err := pageTemplateContent(templateID)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.documentTypeLocked() != document.MDZ {
		return fmt.Errorf("ページテンプレートは通常MDZで使用してください")
	}
	return a.addPageLocked(name, text)
}


type DocumentTemplate struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

var builtInDocumentTemplates = []DocumentTemplate{
	{ID:"document.memo", Kind:"mdz", Title:"メモ"},
	{ID:"document.meeting", Kind:"mdz", Title:"議事録"},
	{ID:"document.spec", Kind:"mdz", Title:"仕様書"},
}

func (a *App) DocumentTemplates() []DocumentTemplate {
	return append([]DocumentTemplate(nil), builtInDocumentTemplates...)
}

func documentTemplateContent(id string) (string, error) {
	switch id {
	case "document.memo":
		return "# メモ\n\n## 内容\n\n", nil
	case "document.meeting":
		return "# 議事録\n\n- 日時:\n- 参加者:\n\n## 議題\n\n## 決定事項\n\n## TODO\n\n", nil
	case "document.spec":
		return "# 仕様書\n\n## 目的\n\n## 要件\n\n## 仕様\n\n## 備考\n\n", nil
	default:
		return "", fmt.Errorf("文書テンプレートが見つかりません")
	}
}

func (a *App) NewDocumentFromTemplate(templateID string) (bool, error) {
	if !a.discardAllowed() {
		return false, nil
	}
	text, err := documentTemplateContent(templateID)
	if err != nil {
		return false, err
	}
	doc, err := newDocument("mdz")
	if err != nil {
		return false, err
	}
	doc.Files[doc.Entry] = []byte(text)
	a.mu.Lock()
	defer a.mu.Unlock()
	session, err := workspace.New(a.base, doc, "", true)
	if err != nil {
		return false, err
	}
	a.adoptLocked(session)
	return true, nil
}
