package main

import "github.com/hatolife/HatoNote/internal/bundle"

type DocumentType string

const (
	DocumentTypeMarkdown DocumentType = "markdown"
	DocumentTypeMDZ      DocumentType = "mdz"
	DocumentTypeMdBook   DocumentType = "mdbook"
	DocumentTypeSlides   DocumentType = "slides"
)

type DocumentCapabilities struct {
	MultiplePages   bool `json:"multiplePages"`
	EmbeddedAssets  bool `json:"embeddedAssets"`
	HeadingTOC      bool `json:"headingToc"`
	MdBookPreview   bool `json:"mdbookPreview"`
	Presentation    bool `json:"presentation"`
	ExternalEditor  bool `json:"externalEditor"`
}

// DetectDocumentType はファイル名と文書構造から文書種別を判定します。
func DetectDocumentType(filename string, doc *bundle.Document) DocumentType {
	if bundle.IsMarkdown(filename) { return DocumentTypeMarkdown; }
	if doc == nil { return ""; }
	if doc.Mode == "slides" {
		return DocumentTypeSlides
	}
	if _, ok := doc.Files["slides.json"]; ok {
		return DocumentTypeSlides
	}
	if _, ok := doc.Files["book.toml"]; ok {
		return DocumentTypeMdBook
	}
	return DocumentTypeMDZ
}

// Capabilities は文書種別ごとに利用可能な共通能力を返します。
func (t DocumentType) Capabilities() DocumentCapabilities {
	switch t {
	case DocumentTypeMarkdown:
		return DocumentCapabilities{HeadingTOC: true, ExternalEditor: true}
	case DocumentTypeMDZ:
		return DocumentCapabilities{MultiplePages: true, EmbeddedAssets: true, HeadingTOC: true, ExternalEditor: true}
	case DocumentTypeMdBook:
		return DocumentCapabilities{MultiplePages: true, EmbeddedAssets: true, MdBookPreview: true, ExternalEditor: true}
	case DocumentTypeSlides:
		return DocumentCapabilities{MultiplePages: true, EmbeddedAssets: true, Presentation: true, ExternalEditor: true}
	default:
		return DocumentCapabilities{}
	}
}
