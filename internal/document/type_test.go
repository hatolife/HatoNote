package document

import (
	"testing"

	"github.com/hatolife/HatoNote/internal/bundle"
)

func TestDetect(t *testing.T) {
	t.Run("Markdown", func(t *testing.T) {
		if got := Detect("note.md", bundle.New()); got != Markdown {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("MDZ", func(t *testing.T) {
		if got := Detect("note.mdz", bundle.New()); got != MDZ {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("mdBook", func(t *testing.T) {
		doc := bundle.New()
		doc.Files["book.toml"] = []byte("[book]\n")
		if got := Detect("book.mdz", doc); got != MdBook {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("Slides", func(t *testing.T) {
		doc := bundle.New()
		doc.Files["slides.json"] = []byte("{}")
		if got := Detect("slides.mdz", doc); got != Slides {
			t.Fatalf("got %q", got)
		}
	})
}

func TestCapabilities(t *testing.T) {
	if Markdown.Capabilities().MultiplePages {
		t.Fatal("Markdown must not support multiple pages")
	}
	if !MDZ.Capabilities().EmbeddedAssets {
		t.Fatal("MDZ must support embedded assets")
	}
	if !MdBook.Capabilities().MdBookPreview {
		t.Fatal("mdBook must support mdBook preview")
	}
	if !Slides.Capabilities().Presentation {
		t.Fatal("slides must support presentation")
	}
	if !Markdown.Capabilities().WYSIWYGEditor || !MDZ.Capabilities().WYSIWYGEditor || !Slides.Capabilities().WYSIWYGEditor {
		t.Fatal("Markdown, MDZ and slides must support WYSIWYG")
	}
	if MdBook.Capabilities().WYSIWYGEditor {
		t.Fatal("mdBook must not support WYSIWYG yet")
	}
}
