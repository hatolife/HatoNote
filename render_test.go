package main

import (
	"strings"
	"testing"
)

func TestRenderDoesNotExecuteDocumentHTML(t *testing.T) {
	app := &App{}
	for _, input := range []string{`<script>alert(1)</script>`, `<img src=x onerror=alert(1)>`, `[click](javascript:alert%281%29)`} {
		result, err := app.Render(input)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(result, "<script") || strings.Contains(result, "onerror=") || strings.Contains(result, `href="javascript:`) {
			t.Fatalf("unsafe output: %s", result)
		}
	}
}


func TestRenderAllowsSafeRichFormatting(t *testing.T) {
	app := &App{}
	input := "<details open><summary>詳細</summary>\n\n<span style=\"color:#ff0000;font-size:20px;position:fixed\" onclick=\"alert(1)\">赤字</span>\n\n</details>"
	result, err := app.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"<details open", "<summary>詳細</summary>", "color:#ff0000", "font-size:20px"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("missing %q: %s", expected, result)
		}
	}
	if strings.Contains(result, "position:") || strings.Contains(result, "onclick=") {
		t.Fatalf("unsafe style or event survived: %s", result)
	}
}

func TestRenderDecoratesManagedReferences(t *testing.T) {
	app := &App{}
	input := "# 本文\n\n引用[@ref-001]\n\n<!-- HATONOTE_REFERENCES_BEGIN -->\n## 参考文献\n\n1. <!-- HATONOTE_REF:ref-001 --> [資料](https://example.com)\n<!-- HATONOTE_REFERENCES_END -->\n"
	result, err := app.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"data-hatonote-citation=\"ref-001\"", "href=\"#hatonote-references\"", ">[1]</a>", "class=\"hatonote-references\"", "資料"} {
		if !strings.Contains(result, expected) {
			t.Fatalf("missing %q: %s", expected, result)
		}
	}
	if strings.Contains(result, "HATONOTE_REF:") {
		t.Fatalf("reference metadata leaked into display: %s", result)
	}
}

func TestRenderLeavesCitationSyntaxInsideCode(t *testing.T) {
	app := &App{}
	result, err := app.Render("```text\n[@ref-001]\n```\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "[@ref-001]") || strings.Contains(result, "data-hatonote-citation") {
		t.Fatalf("citation inside code was decorated: %s", result)
	}
}
