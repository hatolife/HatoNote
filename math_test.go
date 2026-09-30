package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/settings"
)

func TestRenderMathWithConfiguredKaTeX(t *testing.T) {
	script := writeTestExecutable(t, "katex", "args=\" $* \"\ncase \"$args\" in *\" -F mathml \"*) ;; *) exit 5 ;; esac\ninput=$(cat)\nprintf '<span class=\"katex\"><math xmlns=\"http://www.w3.org/1998/Math/MathML\"><mtext>%s</mtext></math></span>' \"$input\"")
	cfg := settings.Default()
	cfg.KatexPath = script
	value, err := renderMath(`c = \\sqrt{a^2+b^2}`, false, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(value, "<math") || !strings.Contains(value, "sqrt") {
		t.Fatalf("mathml = %q", value)
	}
}

func TestRenderMathDisplayMode(t *testing.T) {
	script := writeTestExecutable(t, "katex", "args=\" $* \"\ncase \"$args\" in *\" -d \"*) ;; *) exit 6 ;; esac\ncat >/dev/null\nprintf '<span class=\"katex\"><math xmlns=\"http://www.w3.org/1998/Math/MathML\" display=\"block\"><mn>1</mn></math></span>'")
	cfg := settings.Default()
	cfg.KatexPath = script
	value, err := renderMath("1", true, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(value, `display="block"`) {
		t.Fatalf("display mathml = %q", value)
	}
}

func TestRenderMathRequiresKaTeX(t *testing.T) {
	cfg := settings.Default()
	cfg.KatexPath = filepath.Join(t.TempDir(), "missing-katex")
	if _, err := renderMath("x", false, cfg); err == nil {
		t.Fatal("missing KaTeX dependency accepted")
	}
}

func TestRenderMathRejectsNonMathMLOutput(t *testing.T) {
	script := writeTestExecutable(t, "katex", "cat >/dev/null\nprintf not-mathml")
	cfg := settings.Default()
	cfg.KatexPath = script
	if _, err := renderMath("x", false, cfg); err == nil {
		t.Fatal("non-MathML output accepted")
	}
}
