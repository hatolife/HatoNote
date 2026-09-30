package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckDiagramDependenciesWithConfiguredTools(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("portable executable fixture uses shell scripts")
	}
	mermaid := writeTestExecutable(t, "mmdc", "exit 0")
	java := writeTestExecutable(t, "java", "exit 0")
	katex := writeTestExecutable(t, "katex", "exit 0")
	jar := filepath.Join(t.TempDir(), "plantuml.jar")
	if err := os.WriteFile(jar, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	app := &App{}
	got := app.CheckDiagramDependencies(mermaid, java, jar, katex)
	if len(got) != 4 {
		t.Fatalf("dependencies = %+v", got)
	}
	found := map[string]bool{}
	for _, dependency := range got {
		found[dependency.Name] = dependency.Found
	}
	for _, name := range []string{"Mermaid", "Java", "PlantUML JAR", "KaTeX"} {
		if !found[name] {
			t.Fatalf("%s was not detected: %+v", name, got)
		}
	}
}

func TestCheckDiagramDependenciesReportsMissingTools(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	app := &App{}
	got := app.CheckDiagramDependencies(missing, missing, missing+".jar", missing)
	if len(got) != 4 {
		t.Fatalf("dependencies = %+v", got)
	}
	for _, dependency := range got {
		if dependency.Found {
			t.Fatalf("missing dependency was detected: %+v", dependency)
		}
		if dependency.Message == "" {
			t.Fatalf("missing dependency has no message: %+v", dependency)
		}
	}
}


func TestResolveDiagramExecutableRequiresExecutableBit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission semantics")
	}
	file := filepath.Join(t.TempDir(), "renderer")
	if err := os.WriteFile(file, []byte("#!/bin/sh\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := resolveDiagramExecutable(file, "missing"); got != "" {
		t.Fatalf("non-executable file accepted: %q", got)
	}
	if err := os.Chmod(file, 0700); err != nil {
		t.Fatal(err)
	}
	if got := resolveDiagramExecutable(file, "missing"); got == "" {
		t.Fatal("executable file rejected")
	}
}
