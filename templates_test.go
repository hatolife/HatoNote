package main

import (
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/settings"
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
