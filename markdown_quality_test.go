package main

import (
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/settings"
)

func TestMarkdownQualityUsesSettings(t *testing.T) {
	app := &App{cfg: settings.Default()}
	formatted, err := app.FormatMarkdown("# Title   \n\n\n\nText   ")
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "# Title\n\n\nText\n" {
		t.Fatalf("formatted = %q", formatted)
	}
	diagnostics, err := app.LintMarkdown("# H1  \n### H3")
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string]bool{}
	for _, diagnostic := range diagnostics {
		rules[diagnostic.Rule] = true
	}
	for _, rule := range []string{"trailing-whitespace", "heading-step", "final-newline"} {
		if !rules[rule] {
			t.Fatalf("missing %s in %+v", rule, diagnostics)
		}
	}
}

func TestMarkdownQualitySettingsAreCustomizable(t *testing.T) {
	cfg := settings.Default()
	cfg.FormatTrimTrailingWhitespace = false
	cfg.FormatFinalNewline = false
	cfg.FormatMaxBlankLines = 10
	cfg.LintTrailingWhitespace = false
	cfg.LintHeadingStep = false
	cfg.LintFinalNewline = false
	cfg.LintLongLines = true
	cfg.LintMaxLineLength = 40
	app := &App{cfg: cfg}
	input := "text   "
	formatted, err := app.FormatMarkdown(input)
	if err != nil {
		t.Fatal(err)
	}
	if formatted != input {
		t.Fatalf("custom formatter = %q", formatted)
	}
	diagnostics, err := app.LintMarkdown(strings.Repeat("x", 41))
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 1 || diagnostics[0].Rule != "line-length" {
		t.Fatalf("custom lint = %+v", diagnostics)
	}
}
