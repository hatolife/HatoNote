package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/bundle"
)

func TestBuildSingleHTMLDocument(t *testing.T) {
	doc := bundle.New()
	doc.Entry = "index.md"
	doc.Files["index.md"] = []byte("# One\n\n[next](second.md#two)\n\n![image](images/test.png)\n")
	doc.Files["second.md"] = []byte("# Two\n\nBody\n")
	doc.Files["images/test.png"] = []byte{0x89, 0x50, 0x4e, 0x47}
	doc.Manifest[bundle.PageOrderKey] = json.RawMessage(`["index.md","second.md"]`)
	data, err := buildSingleHTML(doc, "Export Test")
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	if strings.Count(html, `class="hatonote-page"`) != 2 {
		t.Fatalf("page count mismatch: %s", html)
	}
	if !strings.Contains(html, `href="#hatonote-page-2-two"`) {
		t.Fatal("internal Markdown link was not rewritten")
	}
	if !strings.Contains(html, "data:image/png;base64,") {
		t.Fatal("image was not embedded")
	}
	if !strings.Contains(html, "<title>Export Test</title>") {
		t.Fatal("title missing")
	}
}

func TestBuildSingleHTMLSlides(t *testing.T) {
	doc, err := bundle.NewSlides()
	if err != nil {
		t.Fatal(err)
	}
	deck, err := doc.Deck()
	if err != nil {
		t.Fatal(err)
	}
	deck.PageNumberEnabled = true
	deck.PageNumberStart = 3
	deck.PageNumberPosition = "top-left"
	deck.Slides[0].HidePageNumber = true
	encoded, err := json.Marshal(deck)
	if err != nil {
		t.Fatal(err)
	}
	doc.Files["slides.json"] = encoded
	data, err := buildSingleHTML(doc, "Slides")
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	if strings.Count(html, `class="export-slide"`) != len(deck.Slides) {
		t.Fatal("slide count mismatch")
	}
	if strings.Count(html, `class="slide-page-number"`) != len(deck.Slides)-1 {
		t.Fatal("hidden slide page number was rendered")
	}
	if !strings.Contains(html, `data-position="top-left">4</div>`) {
		t.Fatal("slide start number or position was not preserved")
	}
	if !strings.Contains(html, "data:image/svg+xml;base64,") {
		t.Fatal("slide image was not embedded")
	}
}
