package main

import (
	"testing"

	"github.com/hatolife/HatoNote/internal/bundle"
)

func TestDetectDocumentType(t *testing.T) {
	t.Run("Markdown", func(t *testing.T) {
		if got := DetectDocumentType("note.md", bundle.New()); got != DocumentTypeMarkdown { t.Fatalf("got %q", got); }
	})
	t.Run("MDZ", func(t *testing.T) {
		if got := DetectDocumentType("note.mdz", bundle.New()); got != DocumentTypeMDZ { t.Fatalf("got %q", got); }
	})
	t.Run("mdBook", func(t *testing.T) {
		doc := bundle.New()
		doc.Files["book.toml"] = []byte("[book]\n")
		if got := DetectDocumentType("book.mdz", doc); got != DocumentTypeMdBook { t.Fatalf("got %q", got); }
	})
	t.Run("Slides", func(t *testing.T) {
		doc := bundle.New()
		doc.Mode = "slides"
		if got := DetectDocumentType("slides.mdz", doc); got != DocumentTypeSlides { t.Fatalf("got %q", got); }
	})
}

func TestDocumentTypeCapabilities(t *testing.T) {
	if DocumentTypeMarkdown.Capabilities().MultiplePages { t.Fatal("Markdown must not support multiple pages"); }
	if !DocumentTypeMDZ.Capabilities().EmbeddedAssets { t.Fatal("MDZ must support embedded assets"); }
	if !DocumentTypeMdBook.Capabilities().MdBookPreview { t.Fatal("mdBook must support mdBook preview"); }
	if !DocumentTypeSlides.Capabilities().Presentation { t.Fatal("slides must support presentation"); }
}
