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

	templates := app.PageTemplates()
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
	templates := app.DocumentTemplates()
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
	if _, err := documentTemplateContent("missing"); err == nil {
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
	for _, template := range app.DocumentTemplates() {
		if template.ID == "" || template.Kind == "" || template.Title == "" {
			t.Fatalf("template metadata = %+v", template)
		}
		if template.Kind != "mdz" {
			t.Fatalf("unexpected template kind: %+v", template)
		}
	}
}
