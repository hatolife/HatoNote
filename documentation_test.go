package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var specificationIDPattern = regexp.MustCompile(`^[A-Z][A-Z0-9-]*-[0-9]{3}:`)
var emphasisPattern = regexp.MustCompile(`(^|[[:space:]（(])([*][^*\n]+[*]|_[^_\n]+_)($|[[:space:]）).,、。])`)

// TestSpecificationDocuments は仕様書が条件列挙の形式を維持していることを確認します。
func TestSpecificationDocuments(t *testing.T) {
	err := filepath.WalkDir("docs", func(path string, entry os.DirEntry, err error) error {
		if err != nil { return err; }
		if entry.IsDir() || filepath.Ext(path) != ".md" { return nil; }
		data, err := os.ReadFile(path)
		if err != nil { return err; }
		checkSpecificationDocument(t, path, string(data))
		return nil
	})
	if err != nil { t.Fatal(err); }
}

// checkSpecificationDocument は仕様書の本文を機械判定可能な形式に限定します。
func checkSpecificationDocument(t *testing.T, path, text string) {
	t.Helper()
	inCode := false
	for index, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode || trimmed == "" || strings.HasPrefix(trimmed, "#") { continue; }
		if strings.HasPrefix(trimmed, ">") { t.Errorf("%s:%d: 引用記法は使用しません", path, index+1); }
		if strings.Contains(trimmed, "**") { t.Errorf("%s:%d: 太字記法は使用しません", path, index+1); }
		if emphasisPattern.MatchString(stripInlineCode(trimmed)) { t.Errorf("%s:%d: 斜体記法は使用しません", path, index+1); }
		if !strings.HasPrefix(trimmed, "- ") {
			t.Errorf("%s:%d: 仕様は箇条書きの条件として記述します", path, index+1)
			continue
		}
		if !specificationIDPattern.MatchString(strings.TrimPrefix(trimmed, "- ")) {
			t.Errorf("%s:%d: 箇条書きには仕様IDを付けます", path, index+1)
		}
	}
}

// stripInlineCode は仕様値として記述したインラインコードを装飾判定から除外します。
func stripInlineCode(line string) string {
	var result strings.Builder
	inCode := false
	for _, r := range line {
		if r == '`' {
			inCode = !inCode
			continue
		}
		if !inCode { result.WriteRune(r); }
	}
	return result.String()
}
