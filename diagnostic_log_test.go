package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiagnosticLogReadsCurrentFile(t *testing.T) {
	app := &App{base: t.TempDir()}
	if text, err := app.DiagnosticLog(); err != nil || text != "" {
		t.Fatalf("missing log text=%q err=%v", text, err)
	}
	const want = "latest diagnostic log\n"
	if err := os.WriteFile(filepath.Join(app.base, "HatoNote.log"), []byte(want), 0600); err != nil {
		t.Fatal(err)
	}
	text, err := app.DiagnosticLog()
	if err != nil {
		t.Fatal(err)
	}
	if text != want {
		t.Fatalf("log=%q want=%q", text, want)
	}
}
