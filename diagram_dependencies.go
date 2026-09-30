package main

import (
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func resolveDiagramExecutable(configured, fallback string) string {
	configured = strings.TrimSpace(configured)
	if configured == "" {
		found, _ := exec.LookPath(fallback)
		return found
	}
	if found, err := exec.LookPath(configured); err == nil {
		return found
	}
	if info, err := os.Stat(configured); err == nil && !info.IsDir() {
		if goruntime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
			return ""
		}
		if absolute, absErr := filepath.Abs(configured); absErr == nil {
			return absolute
		}
		return configured
	}
	return ""
}

func dependencyForExecutable(name, configured, fallback string) Dependency {
	found := resolveDiagramExecutable(configured, fallback)
	dependency := Dependency{Name:name, Path:found, Found:found != ""}
	if dependency.Found {
		if absolute, err := filepath.Abs(found); err == nil {
			dependency.Path = absolute
		}
		dependency.Message = "検出済み：" + dependency.Path
	} else if strings.TrimSpace(configured) == "" {
		dependency.Message = fallback + " がPATHに見つかりません。実行ファイルを指定してください。"
	} else {
		dependency.Message = "指定した実行ファイルが見つかりません：" + configured
	}
	return dependency
}

// CheckDiagramDependencies はMermaidとPlantUMLの描画依存を設定保存前に確認します。
func (a *App) CheckDiagramDependencies(mermaidPath, javaPath, plantUMLJar, katexPath string) []Dependency {
	mermaid := dependencyForExecutable("Mermaid", mermaidPath, "mmdc")
	java := dependencyForExecutable("Java", javaPath, "java")
	katex := dependencyForExecutable("KaTeX", katexPath, "katex")
	plantuml := Dependency{Name:"PlantUML JAR", Path:strings.TrimSpace(plantUMLJar)}
	if plantuml.Path == "" {
		plantuml.Message = "未指定です。PlantUMLを使用する場合はplantuml.jarを指定してください。"
	} else if info, err := os.Stat(plantuml.Path); err == nil && !info.IsDir() && strings.EqualFold(filepath.Ext(plantuml.Path), ".jar") {
		plantuml.Found = true
		if absolute, absErr := filepath.Abs(plantuml.Path); absErr == nil {
			plantuml.Path = absolute
		}
		plantuml.Message = "検出済み：" + plantuml.Path
	} else {
		plantuml.Message = "PlantUML JARを確認できません：" + plantuml.Path
	}
	return []Dependency{mermaid, java, plantuml, katex}
}

func (a *App) ChooseMermaid() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title:"mmdcを指定", Filters:[]runtime.FileFilter{{DisplayName:"実行ファイル", Pattern:"*.exe;*.cmd;*.bat"}}})
}

func (a *App) ChooseJava() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title:"javaを指定", Filters:[]runtime.FileFilter{{DisplayName:"実行ファイル", Pattern:"*.exe;*.cmd;*.bat"}}})
}

func (a *App) ChoosePlantUMLJar() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title:"plantuml.jarを指定", Filters:[]runtime.FileFilter{{DisplayName:"Java Archive", Pattern:"*.jar"}}})
}

func (a *App) ChooseKatex() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title:"katexを指定", Filters:[]runtime.FileFilter{{DisplayName:"実行ファイル", Pattern:"*.exe;*.cmd;*.bat"}}})
}
