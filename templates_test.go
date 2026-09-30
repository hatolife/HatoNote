package main

import (
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/settings"
	"github.com/hatolife/HatoNote/internal/document"
	"github.com/hatolife/HatoNote/internal/workspace"
)

func TestBuiltInPageTemplates(t *testing.T) {
	app := &App{base: t.TempDir(), cfg: settings.Default()}
	doc, err := newDocument("mdz")
	if err != nil {
		t.Fatal(err)
	}
	app.session, err = workspace.New(app.base, doc, "", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.shutdown)

	templates, err := app.PageTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 4 {
		t.Fatalf("templates = %+v", templates)
	}
	for _, template := range templates {
		if template.ID == "" || template.Title == "" || template.Content != "" {
			t.Fatalf("template metadata = %+v", template)
		}
	}

	if err := app.AddPageFromTemplate("meeting", "meeting"); err != nil {
		t.Fatal(err)
	}
	text := string(app.session.Doc.Files["meeting.md"])
	if !strings.Contains(text, "# 議事録") || !strings.Contains(text, "## TODO") {
		t.Fatalf("meeting template = %q", text)
	}
	pages := app.State().Pages
	if pages[len(pages)-1] != "meeting.md" {
		t.Fatalf("pages = %v", pages)
	}
	if err := app.AddPageFromTemplate("meeting", "memo"); err == nil {
		t.Fatal("duplicate page name accepted")
	}
	if err := app.AddPageFromTemplate("other", "missing"); err == nil {
		t.Fatal("unknown template accepted")
	}
}

func TestPageTemplateOnlyForNormalMDZ(t *testing.T) {
	app := &App{base: t.TempDir(), cfg: settings.Default()}
	doc, err := newDocument("mdbook")
	if err != nil {
		t.Fatal(err)
	}
	app.session, err = workspace.New(app.base, doc, "", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.shutdown)
	if err := app.AddPageFromTemplate("memo", "memo"); err == nil {
		t.Fatal("mdBook accepted normal MDZ page template")
	}
}


func TestDocumentTemplates(t *testing.T) {
	app := &App{base:t.TempDir(), cfg:settings.Default()}
	templates, err := app.DocumentTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 3 {
		t.Fatalf("document templates = %+v", templates)
	}
	ok, err := app.NewDocumentFromTemplate("document.meeting")
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	t.Cleanup(app.shutdown)
	if app.State().DocumentType != document.MDZ {
		t.Fatalf("document type = %q", app.State().DocumentType)
	}
	if app.session.Doc.Entry != "本文.md" || !strings.Contains(string(app.session.Doc.Files["本文.md"]), "# 議事録") {
		t.Fatalf("document template = %q", app.session.Doc.Files["本文.md"])
	}
	if _, err := app.documentTemplateContent("missing"); err == nil {
		t.Fatal("unknown document template accepted")
	}
}


func TestBlankDocumentKinds(t *testing.T) {
	for _, kind := range []string{"markdown", "mdz", "mdbook", "slides"} {
		doc, err := newBlankDocument(kind)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if doc.Entry == "" || len(doc.Files) == 0 {
			t.Fatalf("%s blank document is incomplete: %+v", kind, doc)
		}
	}
	markdown, err := newBlankDocument("markdown")
	if err != nil {
		t.Fatal(err)
	}
	if got := document.Detect("", markdown); got != document.Markdown {
		t.Fatalf("blank markdown type = %q", got)
	}
	if string(markdown.Files[markdown.Entry]) != "" {
		t.Fatalf("blank markdown content = %q", markdown.Files[markdown.Entry])
	}
}

func TestDocumentTemplatesHaveKinds(t *testing.T) {
	app := &App{}
	templates, err := app.DocumentTemplates()
	if err != nil {
		t.Fatal(err)
	}
	for _, template := range templates {
		if template.ID == "" || template.Kind == "" || template.Title == "" {
			t.Fatalf("template metadata = %+v", template)
		}
		if template.Kind != "mdz" {
			t.Fatalf("unexpected template kind: %+v", template)
		}
	}
}


func TestUserDefinedTemplates(t *testing.T) {
	app := &App{base:t.TempDir(), cfg:settings.Default()}
	page, err := app.SaveUserTemplate(UserTemplate{
		Scope:"page",
		Kind:"mdz",
		Title:"レビュー",
		Content:"# レビュー\n\n## 指摘\n\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.ID == "" || !strings.HasPrefix(page.ID, "user.") {
		t.Fatalf("page template id = %q", page.ID)
	}
	documentTemplate, err := app.SaveUserTemplate(UserTemplate{
		Scope:"document",
		Kind:"mdz",
		Title:"設計メモ",
		Content:"# 設計メモ\n\n## 方針\n\n",
	})
	if err != nil {
		t.Fatal(err)
	}

	pageTemplates, err := app.PageTemplates()
	if err != nil {
		t.Fatal(err)
	}
	foundPage := false
	for _, template := range pageTemplates {
		if template.ID == page.ID && template.Title == "レビュー" {
			foundPage = true
		}
	}
	if !foundPage {
		t.Fatalf("user page template missing: %+v", pageTemplates)
	}

	documentTemplates, err := app.DocumentTemplates()
	if err != nil {
		t.Fatal(err)
	}
	foundDocument := false
	for _, template := range documentTemplates {
		if template.ID == documentTemplate.ID && template.Title == "設計メモ" && template.Kind == "mdz" {
			foundDocument = true
		}
	}
	if !foundDocument {
		t.Fatalf("user document template missing: %+v", documentTemplates)
	}

	doc, err := newDocument("mdz")
	if err != nil {
		t.Fatal(err)
	}
	app.session, err = workspace.New(app.base, doc, "", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.shutdown)
	if err := app.AddPageFromTemplate("review", page.ID); err != nil {
		t.Fatal(err)
	}
	if got := string(app.session.Doc.Files["review.md"]); !strings.Contains(got, "# レビュー") {
		t.Fatalf("page content = %q", got)
	}

	items, err := app.UserTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("user templates = %+v", items)
	}
	page.Title = "レビュー改"
	page.Content = "# レビュー改\n"
	updated, err := app.SaveUserTemplate(page)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "レビュー改" {
		t.Fatalf("updated = %+v", updated)
	}
	if err := app.DeleteUserTemplate(documentTemplate.ID); err != nil {
		t.Fatal(err)
	}
	items, err = app.UserTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != page.ID {
		t.Fatalf("remaining templates = %+v", items)
	}
}

func TestUserTemplateValidation(t *testing.T) {
	app := &App{base:t.TempDir(), cfg:settings.Default()}
	for _, template := range []UserTemplate{
		{Scope:"page", Kind:"mdz", Title:""},
		{Scope:"page", Kind:"slides", Title:"bad"},
		{Scope:"document", Kind:"mdbook", Title:"bad"},
		{Scope:"other", Kind:"mdz", Title:"bad"},
	} {
		if _, err := app.SaveUserTemplate(template); err == nil {
			t.Fatalf("invalid template accepted: %+v", template)
		}
	}
	if err := app.DeleteUserTemplate("memo"); err == nil {
		t.Fatal("built-in template deletion accepted")
	}
}
