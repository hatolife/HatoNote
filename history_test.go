package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hatolife/HatoNote/internal/settings"
	"github.com/hatolife/HatoNote/internal/workspace"
)

func TestHistoryDiffDetectsMarkdownChanges(t *testing.T) {
	base := t.TempDir()
	doc, err := newDocument("mdz")
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Put("removed.md", []byte("# Removed\n")); err != nil {
		t.Fatal(err)
	}
	app := &App{base: base, cfg: settings.Default()}
	app.session, err = workspace.New(base, doc, "", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.shutdown)
	filename := filepath.Join(t.TempDir(), "history.mdz")
	if err := app.session.Save(base, filename, "test", app.cfg); err != nil {
		t.Fatal(err)
	}
	named, err := app.CreateNamedVersion("baseline")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Update("本文.md", "# Changed\n"); err != nil {
		t.Fatal(err)
	}
	if err := app.session.Put("added.md", []byte("# Added\n")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(app.session.Content(), "removed.md")); err != nil {
		t.Fatal(err)
	}
	diff, err := app.HistoryDiff(named.ID)
	if err != nil {
		t.Fatal(err)
	}
	status := map[string]string{}
	pages := map[string]HistoryPageDiff{}
	for _, page := range diff {
		status[page.Page] = page.Status
		pages[page.Page] = page
	}
	if status["本文.md"] != "modified" || status["added.md"] != "added" || status["removed.md"] != "deleted" {
		t.Fatalf("diff = %+v", diff)
	}
	for _, name := range []string{"本文.md", "added.md", "removed.md"} {
		if len(pages[name].Lines) == 0 {
			t.Fatalf("%s has no line diff: %+v", name, pages[name])
		}
	}
	containsStatus := func(lines []HistoryLineDiff, want string) bool {
		for _, line := range lines {
			if line.Status == want {
				return true
			}
		}
		return false
	}
	if !containsStatus(pages["本文.md"].Lines, "modified") ||
		!containsStatus(pages["added.md"].Lines, "added") ||
		!containsStatus(pages["removed.md"].Lines, "deleted") {
		t.Fatalf("line diff statuses = %+v", pages)
	}
}

func TestHistoryDiffRequiresSavedDocument(t *testing.T) {
	doc, err := newDocument("mdz")
	if err != nil {
		t.Fatal(err)
	}
	app := &App{base: t.TempDir(), cfg: settings.Default()}
	app.session, err = workspace.New(app.base, doc, "", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.shutdown)
	if _, err := app.HistoryDiff("named/missing"); err == nil {
		t.Fatal("history diff accepted unsaved document")
	}
}
