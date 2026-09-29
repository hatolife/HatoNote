package main

import (
	"bytes"
	"fmt"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

func renderMarkdown(text string) (string, error) {
	if len(text) > bundle.MaxFile {
		return "", fmt.Errorf("本文が大きすぎます")
	}
	var output bytes.Buffer
	md := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	if err := md.Convert([]byte(text), &output); err != nil {
		return "", err
	}
	return output.String(), nil
}

// Render は生HTMLと危険なURLを許可せずにMarkdownを変換します。
func (a *App) Render(text string) (string, error) {
	return renderMarkdown(text)
}
