package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLimitsAndDefaults(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatal(err)
	}
	if Default().InitMode != "custom" {
		t.Fatalf("default init mode = %q", Default().InitMode)
	}
	if !Default().ShowMarkdownCheatsheet {
		t.Fatal("Markdown cheatsheet must be visible by default")
	}
	for _, editor := range []string{"builtin", "wysiwyg", "neovim"} {
		s := Default()
		s.Editor = editor
		if err := s.Validate(); err != nil {
			t.Fatalf("editor %q rejected: %v", editor, err)
		}
	}
	{
		s := Default()
		s.Editor = "unknown"
		if err := s.Validate(); err == nil {
			t.Fatal("unknown editor accepted")
		}
	}
	for _, n := range []int{0, -1, 10001} {
		s := Default()
		s.UndoLevels = n
		if err := s.Validate(); err == nil {
			t.Fatalf("accepted undo limit %d", n)
		}
	}
	for _, path := range []string{"../outside", "/absolute", "C:/absolute", "images\\windows"} {
		s := Default()
		s.ImageDirectory = path
		if err := s.Validate(); err == nil {
			t.Fatalf("accepted path %q", path)
		}
	}
	s := Default()
	s.ImageName = "{unknown}"
	if err := s.Validate(); err == nil {
		t.Fatal("unknown token accepted")
	}
	file := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(file, []byte(`{"undoLevels":25,"imageDirectory":"assets/pasted"}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if s.UndoLevels != 25 || s.AutoSaveSeconds != 60 || s.ImageDirectory != "assets/pasted" {
		t.Fatalf("wrong defaults: %+v", s)
	}
	if !s.FormatTrimTrailingWhitespace || s.FormatMaxBlankLines != 2 || !s.FormatFinalNewline {
		t.Fatalf("wrong formatter defaults: %+v", s)
	}
	if !s.LintTrailingWhitespace || s.LintLongLines || s.LintMaxLineLength != 120 || !s.LintHeadingStep || !s.LintFinalNewline {
		t.Fatalf("wrong lint defaults: %+v", s)
	}
	for _, value := range []int{-1, 11} {
		bad := Default()
		bad.FormatMaxBlankLines = value
		if err := bad.Validate(); err == nil {
			t.Fatalf("accepted format blank line limit %d", value)
		}
	}
	for _, value := range []int{39, 1001} {
		bad := Default()
		bad.LintMaxLineLength = value
		if err := bad.Validate(); err == nil {
			t.Fatalf("accepted lint line length %d", value)
		}
	}
}
