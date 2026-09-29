package bundle

import "testing"

func TestNewSlideDeckDefaults(t *testing.T) {
	deck := NewSlideDeck("資料")
	if deck.Version != 1 || deck.Title != "資料" || deck.Theme != "light" || deck.Aspect != "16:9" {
		t.Fatalf("unexpected deck defaults: %+v", deck)
	}
	if deck.ContentMarginX != 60 || deck.ContentMarginY != 48 {
		t.Fatalf("unexpected margin defaults: %+v", deck)
	}
	if deck.FontFamily != "system" || deck.BodyFontSize != 20 {
		t.Fatalf("unexpected typography defaults: %+v", deck)
	}
	if deck.H1FontSize != 42 || deck.H2FontSize != 28 || deck.H3FontSize != 26 || deck.H4FontSize != 24 || deck.H5FontSize != 22 {
		t.Fatalf("unexpected heading defaults: %+v", deck)
	}
	if deck.PageNumberEnabled || deck.PageNumberPosition != "bottom-right" || deck.PageNumberStart != 1 {
		t.Fatalf("unexpected page number defaults: %+v", deck)
	}
}


func TestSlideDeckRejectsSharedMarkdownFile(t *testing.T) {
	deck := SlideDeck{
		Version: 1,
		Title:   "資料",
		Theme:   "light",
		Aspect:  "16:9",
		Slides: []Slide{
			{ID: "slide-001", File: "slides/page.md", Title: "1", Layout: "standard", FontSize: 20},
			{ID: "slide-002", File: "slides/page.md", Title: "2", Layout: "standard", FontSize: 20},
		},
	}
	if err := deck.Validate(map[string][]byte{"slides/page.md": []byte("# page\n")}); err == nil {
		t.Fatal("multiple slides sharing one Markdown file were accepted")
	}
}
