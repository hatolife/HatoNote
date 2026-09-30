package markdownquality

import (
	"strings"
	"testing"
)

func TestFormatSkipsCodeFences(t *testing.T) {
	input := "# Title   \n\n\n\nText   \n\n\`\`\`text\ncode   \n\n\n\`\`\`\n"
	got := Format(input, FormatOptions{TrimTrailingWhitespace:true, MaxBlankLines:2, FinalNewline:true})
	if !strings.Contains(got, "# Title\n") || !strings.Contains(got, "Text\n") {
		t.Fatalf("outside whitespace not trimmed: %q", got)
	}
	if !strings.Contains(got, "code   \n\n\n") {
		t.Fatalf("code fence content changed: %q", got)
	}
	if strings.Contains(got, "\n\n\nText") {
		t.Fatalf("blank lines were not limited: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Fatal("final newline missing")
	}
}

func TestLintRulesAndFenceExclusion(t *testing.T) {
	text := "# H1  \n### H3\n" + strings.Repeat("x", 20) + "\n\`\`\`\ncode   \n" + strings.Repeat("y", 30) + "\n\`\`\`"
	got := Lint(text, LintOptions{TrailingWhitespace:true, LongLines:true, MaxLineLength:10, HeadingStep:true, FinalNewline:true})
	rules := map[string]int{}
	for _, diagnostic := range got {
		rules[diagnostic.Rule]++
	}
	if rules["trailing-whitespace"] != 1 || rules["heading-step"] != 1 || rules["line-length"] != 1 || rules["final-newline"] != 1 {
		t.Fatalf("diagnostics = %+v", got)
	}
}

func TestFormatOptionsCanDisableChanges(t *testing.T) {
	input := "text   "
	got := Format(input, FormatOptions{TrimTrailingWhitespace:false, MaxBlankLines:10, FinalNewline:false})
	if got != input {
		t.Fatalf("format changed disabled rule: %q", got)
	}
}
