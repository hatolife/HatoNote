package document

import "github.com/hatolife/HatoNote/internal/bundle"

type Type string

const (
	Markdown Type = "markdown"
	MDZ      Type = "mdz"
	MdBook   Type = "mdbook"
	Slides   Type = "slides"
)

type Capabilities struct {
	MultiplePages  bool `json:"multiplePages"`
	EmbeddedAssets bool `json:"embeddedAssets"`
	HeadingTOC     bool `json:"headingToc"`
	MdBookPreview  bool `json:"mdbookPreview"`
	Presentation   bool `json:"presentation"`
	ExternalEditor bool `json:"externalEditor"`
	WYSIWYGEditor bool `json:"wysiwygEditor"`
}

// Detect はファイル名とHatoNote固有の文書構造から文書種別を判定します。
func Detect(filename string, doc *bundle.Document) Type {
	if bundle.IsMarkdown(filename) {
		return Markdown
	}
	if doc == nil {
		return ""
	}
	if doc.HasSlides() {
		return Slides
	}
	if _, ok := doc.Files["book.toml"]; ok {
		return MdBook
	}
	return MDZ
}

func (t Type) Capabilities() Capabilities {
	switch t {
	case Markdown:
		return Capabilities{HeadingTOC: true, ExternalEditor: true, WYSIWYGEditor: true}
	case MDZ:
		return Capabilities{MultiplePages: true, EmbeddedAssets: true, HeadingTOC: true, ExternalEditor: true, WYSIWYGEditor: true}
	case MdBook:
		return Capabilities{MultiplePages: true, EmbeddedAssets: true, MdBookPreview: true, ExternalEditor: true}
	case Slides:
		return Capabilities{MultiplePages: true, EmbeddedAssets: true, Presentation: true, ExternalEditor: true, WYSIWYGEditor: true}
	default:
		return Capabilities{}
	}
}
