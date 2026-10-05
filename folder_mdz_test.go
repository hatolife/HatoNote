package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hatolife/HatoNote/internal/bundle"
)

func TestBuildMDZDocumentFromFolder(t *testing.T) {
	root := t.TempDir()
	writeFolderTestFile(t, root, "README.md", "# Top\n")
	writeFolderTestFile(t, root, "chapters/2.md", "# Two\n")
	writeFolderTestFile(t, root, "chapters/10.md", "# Ten\n")
	writeFolderTestFile(t, root, "images/picture.png", "image")
	writeFolderTestFile(t, root, ".git/config", "ignored")
	writeFolderTestFile(t, root, "manifest.json", "{}")

	doc, info, err := buildMDZDocumentFromFolder(root)
	if err != nil {
		t.Fatal(err)
	}
	if info.MarkdownCount != 3 {
		t.Fatalf("MarkdownCount = %d, want 3", info.MarkdownCount)
	}
	if info.FileCount != 4 {
		t.Fatalf("FileCount = %d, want 4", info.FileCount)
	}
	if doc.Entry != "README.md" {
		t.Fatalf("Entry = %q, want README.md", doc.Entry)
	}
	for _, name := range []string{"README.md", "chapters/2.md", "chapters/10.md", "images/picture.png"} {
		if _, ok := doc.Files[name]; !ok {
			t.Fatalf("%s was not imported", name)
		}
	}
	for _, name := range []string{".git/config", "manifest.json"} {
		if _, ok := doc.Files[name]; ok {
			t.Fatalf("%s must not be imported", name)
		}
	}
	var order []string
	if err := json.Unmarshal(doc.Manifest[bundle.PageOrderKey], &order); err != nil {
		t.Fatal(err)
	}
	want := []string{"chapters/2.md", "chapters/10.md", "README.md"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("page order = %#v, want %#v", order, want)
	}
}

func TestBuildMDZDocumentFromFolderRequiresMarkdown(t *testing.T) {
	root := t.TempDir()
	writeFolderTestFile(t, root, "images/picture.png", "image")
	if _, _, err := buildMDZDocumentFromFolder(root); err == nil {
		t.Fatal("Markdownがないフォルダはエラーになる必要があります")
	}
}

func TestPreferredFolderEntryUsesRootIndexBeforeNaturalOrder(t *testing.T) {
	markdowns := []string{"chapters/1.md", "index.md", "README.txt.md"}
	if got := preferredFolderEntry(markdowns); got != "index.md" {
		t.Fatalf("Entry = %q, want index.md", got)
	}
}

func writeFolderTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
