package main

import (
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/settings"
	"github.com/hatolife/HatoNote/internal/workspace"
)

func referenceApp(t *testing.T) *App {
	t.Helper()
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
	return app
}

func TestReferencesRoundTripAndManagedBlock(t *testing.T) {
	app := referenceApp(t)
	reference, err := app.SaveReference(Reference{Title: "資料A", URL: "https://example.com/source", Note: "確認日 2026-09-30"})
	if err != nil {
		t.Fatal(err)
	}
	if reference.ID != "ref-001" {
		t.Fatalf("unexpected id: %s", reference.ID)
	}
	set, err := app.References()
	if err != nil || len(set.Items) != 1 || set.Items[0].Title != "資料A" {
		t.Fatalf("references: %+v, %v", set, err)
	}
	page := app.session.Doc.Entry
	if err := app.Update(page, "# 本文\n\n説明[@ref-001]\n"); err != nil {
		t.Fatal(err)
	}
	updated, err := app.ApplyReferences(page)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"[@ref-001]", "<!-- HATONOTE_REFERENCES_BEGIN -->", "<!-- HATONOTE_REF:ref-001 -->", "[資料A](https://example.com/source)", "確認日 2026-09-30"} {
		if !strings.Contains(updated, expected) {
			t.Fatalf("missing %q: %s", expected, updated)
		}
	}
	if err := app.DeleteReference(reference.ID); err == nil {
		t.Fatal("cited reference was deleted")
	}
	if err := app.Update(page, "# 本文\n"); err != nil {
		t.Fatal(err)
	}
	updated, err = app.ApplyReferences(page)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(updated, "HATONOTE_REFERENCES") {
		t.Fatalf("unused managed block remained: %s", updated)
	}
	if err := app.DeleteReference(reference.ID); err != nil {
		t.Fatal(err)
	}
	set, err = app.References()
	if err != nil || len(set.Items) != 0 {
		t.Fatalf("reference was not deleted: %+v, %v", set, err)
	}
}

func TestReferencesAreStoredInManifestExtension(t *testing.T) {
	app := referenceApp(t)
	if _, err := app.SaveReference(Reference{Title: "資料"}); err != nil {
		t.Fatal(err)
	}
	raw := app.session.Doc.Manifest[referenceManifestKey]
	if !strings.Contains(string(raw), "ref-001") || !strings.Contains(string(raw), "資料") {
		t.Fatalf("reference metadata was not stored in manifest: %s", raw)
	}
}

func TestCitationIDsIgnoreFencedCode(t *testing.T) {
	text := "本文[@ref-001]\n\n```text\n[@ref-002]\n```\n\n末尾[@ref-003]\n"
	ids := citationIDs(text)
	if strings.Join(ids, ",") != "ref-001,ref-003" {
		t.Fatalf("ids: %v", ids)
	}
}
