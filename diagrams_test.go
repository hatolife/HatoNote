package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/settings"
)

func writeTestExecutable(t *testing.T, name, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("diagram command fixture is covered on non-Windows CI")
	}
	filename := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(filename, []byte("#!/bin/sh\nset -eu\n"+body+"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	return filename
}

func decodeDiagramDataURI(t *testing.T, value string) string {
	t.Helper()
	const prefix = "data:image/svg+xml;base64,"
	if !strings.HasPrefix(value, prefix) {
		t.Fatalf("data uri = %q", value)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRenderMermaidWithConfiguredCLI(t *testing.T) {
	script := writeTestExecutable(t, "mmdc", "out=\"\"\nwhile [ \"$#\" -gt 0 ]; do\n if [ \"$1\" = \"-o\" ]; then shift; out=\"$1\"; fi\n shift || true\ndone\nprintf '<svg xmlns=\"http://www.w3.org/2000/svg\"><text>mermaid</text></svg>' > \"$out\"")
	cfg := settings.Default()
	cfg.MermaidPath = script
	value, err := renderDiagram("mermaid", "graph TD; A-->B", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(decodeDiagramDataURI(t, value), ">mermaid<") {
		t.Fatal("mermaid SVG missing")
	}
}

func TestRenderPlantUMLWithConfiguredJava(t *testing.T) {
	script := writeTestExecutable(t, "java", "input=$(cat)\ncase \"$input\" in *\"@startuml\"*\"@enduml\"*) ;; *) exit 7 ;; esac\nprintf '<svg xmlns=\"http://www.w3.org/2000/svg\"><text>plantuml</text></svg>'")
	jar := filepath.Join(t.TempDir(), "plantuml.jar")
	if err := os.WriteFile(jar, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := settings.Default()
	cfg.JavaPath = script
	cfg.PlantUMLJar = jar
	value, err := renderDiagram("puml", "Alice -> Bob: Hello", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(decodeDiagramDataURI(t, value), ">plantuml<") {
		t.Fatal("PlantUML SVG missing")
	}
}

func TestDiagramErrorsWithoutDependencies(t *testing.T) {
	cfg := settings.Default()
	cfg.MermaidPath = filepath.Join(t.TempDir(), "missing")
	if _, err := renderDiagram("mermaid", "graph TD; A-->B", cfg); err == nil {
		t.Fatal("missing Mermaid dependency accepted")
	}
	cfg.JavaPath = filepath.Join(t.TempDir(), "missing-java")
	cfg.PlantUMLJar = filepath.Join(t.TempDir(), "plantuml.jar")
	if _, err := renderDiagram("plantuml", "@startuml\nA -> B\n@enduml", cfg); err == nil {
		t.Fatal("missing Java dependency accepted")
	}
}

func TestNormalizePlantUMLKeepsExplicitStart(t *testing.T) {
	source := "@startmindmap\n* Root\n@endmindmap"
	if got := normalizePlantUML(source); got != source {
		t.Fatalf("explicit start changed: %q", got)
	}
}
