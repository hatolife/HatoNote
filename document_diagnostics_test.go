package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/hatolife/HatoNote/internal/document"
	"github.com/hatolife/HatoNote/internal/settings"
)

func hasDocumentDiagnostic(items []DocumentDiagnostic, kind, page, rule string) bool {
	for _, item := range items {
		if item.Type == kind && item.Page == page && item.Rule == rule {
			return true
		}
	}
	return false
}

func TestDiagnoseDocumentAppliesLintToAllMarkdownPages(t *testing.T) {
	doc := bundle.New()
	doc.Files["a.md"] = []byte("# A  \n")
	doc.Files["b.md"] = []byte("# B")

	items := diagnoseDocument(doc, "sample.mdz", document.MDZ, settings.Default())
	if !hasDocumentDiagnostic(items, "lint", "a.md", "trailing-whitespace") {
		t.Fatal("a.mdのLint結果がありません")
	}
	if !hasDocumentDiagnostic(items, "lint", "b.md", "final-newline") {
		t.Fatal("b.mdのLint結果がありません")
	}
}

func TestDiagnoseDocumentChecksLinksImagesAndUnusedImages(t *testing.T) {
	doc := bundle.New()
	doc.Files["a.md"] = []byte("# A\n\n![used](images/used.png)\n![missing](images/missing.png)\n[missing](missing.md)\n[external](https://example.com/x)\n[mail](mailto:user@example.com)\n[anchor](#a)\n")
	doc.Files["images/used.png"] = []byte("used")
	doc.Files["images/unused.png"] = []byte("unused")

	items := diagnoseDocument(doc, "sample.mdz", document.MDZ, settings.Default())
	if !hasDocumentDiagnostic(items, "broken-image", "a.md", "") {
		t.Fatal("切れた画像が検出されません")
	}
	if !hasDocumentDiagnostic(items, "broken-link", "a.md", "") {
		t.Fatal("切れたリンクが検出されません")
	}
	brokenLinks := 0
	unused := 0
	for _, item := range items {
		if item.Type == "broken-link" {
			brokenLinks++
		}
		if item.Type == "unused-image" {
			unused++
			if item.Message != "未使用画像: images/unused.png" {
				t.Fatalf("未使用画像が不正です: %s", item.Message)
			}
		}
	}
	if brokenLinks != 1 {
		t.Fatalf("外部URLまたはアンカーまで切れたリンクとして検出されています: %d", brokenLinks)
	}
	if unused != 1 {
		t.Fatalf("未使用画像の件数が不正です: %d", unused)
	}
}

func TestDiagnoseDocumentResolvesNestedBundleReferences(t *testing.T) {
	doc := bundle.New()
	doc.Files["chapters/a.md"] = []byte("![used](../images/used.png)\n[next](b.md)\n")
	doc.Files["chapters/b.md"] = []byte("# B\n")
	doc.Files["images/used.png"] = []byte("used")

	items := diagnoseDocument(doc, "sample.mdz", document.MDZ, settings.Default())
	for _, item := range items {
		if item.Type == "broken-link" || item.Type == "broken-image" || item.Type == "unused-image" {
			t.Fatalf("有効な相対参照が誤診断されました: %+v", item)
		}
	}
}

func TestDiagnoseSingleMarkdownUsesSourceDirectory(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "note.md")
	if err := os.WriteFile(filepath.Join(dir, "exists.md"), []byte("# exists\n"), 0600); err != nil {
		t.Fatal(err)
	}
	doc := bundle.New()
	doc.Files["note.md"] = []byte("[ok](exists.md)\n[missing](missing.md)\n")
	doc.Entry = "note.md"

	items := diagnoseDocument(doc, filename, document.Markdown, settings.Default())
	broken := 0
	for _, item := range items {
		if item.Type == "broken-link" {
			broken++
		}
		if item.Type == "unused-image" {
			t.Fatal("単一Markdownで未使用画像診断が実行されました")
		}
	}
	if broken != 1 {
		t.Fatalf("切れたリンクの件数が不正です: %d", broken)
	}
}
