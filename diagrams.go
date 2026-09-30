package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/hatolife/HatoNote/internal/process"
	"github.com/hatolife/HatoNote/internal/settings"
)

const maxDiagramSource = 1 << 20
const maxDiagramSVG = 8 << 20

var diagramCache = struct {
	sync.Mutex
	values map[string]string
}{values: map[string]string{}}

func diagramCacheKey(kind, source string, cfg settings.Settings) string {
	value := strings.Join([]string{kind, source, cfg.MermaidPath, cfg.JavaPath, cfg.PlantUMLJar}, "\x00")
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func diagramCacheGet(key string) (string, bool) {
	diagramCache.Lock()
	defer diagramCache.Unlock()
	value, ok := diagramCache.values[key]
	return value, ok
}

func diagramCachePut(key, value string) {
	diagramCache.Lock()
	defer diagramCache.Unlock()
	if len(diagramCache.values) >= 64 {
		for existing := range diagramCache.values {
			delete(diagramCache.values, existing)
			break
		}
	}
	diagramCache.values[key] = value
}

func diagramDataURI(svg []byte) (string, error) {
	if len(svg) == 0 || len(svg) > maxDiagramSVG {
		return "", fmt.Errorf("生成SVGのサイズが不正です")
	}
	trimmed := bytes.TrimSpace(svg)
	if !bytes.Contains(trimmed, []byte("<svg")) {
		return "", fmt.Errorf("SVGを生成できませんでした")
	}
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(svg), nil
}

func cmdQuote(value string) string {
	return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
}

func diagramCommandContext(ctx context.Context, executable string, args ...string) *exec.Cmd {
	extension := strings.ToLower(filepath.Ext(executable))
	if goruntime.GOOS == "windows" && (extension == ".cmd" || extension == ".bat") {
		parts := []string{"call", cmdQuote(executable)}
		for _, arg := range args {
			parts = append(parts, cmdQuote(arg))
		}
		return process.CommandContext(ctx, "cmd.exe", "/d", "/s", "/c", strings.Join(parts, " "))
	}
	return process.CommandContext(ctx, executable, args...)
}

func readDiagramSVG(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxDiagramSVG+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxDiagramSVG {
		return nil, fmt.Errorf("生成SVGが8MiBを超えています")
	}
	return data, nil
}

func renderMermaid(ctx context.Context, source string, cfg settings.Settings) ([]byte, error) {
	executable := resolveDiagramExecutable(cfg.MermaidPath, "mmdc")
	if executable == "" {
		return nil, fmt.Errorf("Mermaid CLIが見つかりません。設定からmmdcを指定してください")
	}
	dir, err := os.MkdirTemp("", "HatoNote-mermaid-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "diagram.mmd")
	output := filepath.Join(dir, "diagram.svg")
	config := filepath.Join(dir, "config.json")
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(config, []byte("{\"securityLevel\":\"strict\",\"flowchart\":{\"htmlLabels\":false}}"), 0600); err != nil {
		return nil, err
	}
	command := diagramCommandContext(ctx, executable, "-i", input, "-o", output, "-c", config, "-b", "transparent")
	var stderr bytes.Buffer
	command.Stdout = io.Discard
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("Mermaidの描画に失敗しました: %s", message)
	}
	return readDiagramSVG(output)
}

func normalizePlantUML(source string) string {
	if strings.Contains(strings.ToLower(source), "@start") {
		return source
	}
	return "@startuml\n" + source + "\n@enduml\n"
}

func renderPlantUML(ctx context.Context, source string, cfg settings.Settings) ([]byte, error) {
	java := resolveDiagramExecutable(cfg.JavaPath, "java")
	if java == "" {
		return nil, fmt.Errorf("Javaが見つかりません。設定からJavaを指定してください")
	}
	jar := strings.TrimSpace(cfg.PlantUMLJar)
	if jar == "" {
		return nil, fmt.Errorf("PlantUML JARが未指定です")
	}
	info, err := os.Stat(jar)
	if err != nil || info.IsDir() {
		return nil, fmt.Errorf("PlantUML JARを確認できません")
	}
	command := process.CommandContext(ctx, java,
		"-DPLANTUML_SECURITY_PROFILE=SANDBOX",
		"-jar", jar,
		"-tsvg",
		"-pipe",
	)
	command.Stdin = strings.NewReader(normalizePlantUML(source))
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("PlantUMLの描画に失敗しました: %s", message)
	}
	if stdout.Len() > maxDiagramSVG {
		return nil, fmt.Errorf("生成SVGが8MiBを超えています")
	}
	return stdout.Bytes(), nil
}

func renderDiagram(kind, source string, cfg settings.Settings) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", fmt.Errorf("図表ソースが空です")
	}
	if len(source) > maxDiagramSource {
		return "", fmt.Errorf("図表ソースは1MiB以下にしてください")
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "puml" {
		kind = "plantuml"
	}
	if kind != "mermaid" && kind != "plantuml" {
		return "", fmt.Errorf("未対応の図表種類です")
	}
	key := diagramCacheKey(kind, source, cfg)
	if value, ok := diagramCacheGet(key); ok {
		return value, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var svg []byte
	var err error
	switch kind {
	case "mermaid":
		svg, err = renderMermaid(ctx, source, cfg)
	case "plantuml":
		svg, err = renderPlantUML(ctx, source, cfg)
	}
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("図表描画が20秒でタイムアウトしました")
		}
		return "", err
	}
	value, err := diagramDataURI(svg)
	if err != nil {
		return "", err
	}
	diagramCachePut(key, value)
	return value, nil
}

// RenderDiagram はMarkdownコードフェンスの図表をSVG画像へ変換します。
func (a *App) RenderDiagram(kind, source string) (string, error) {
	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()
	return renderDiagram(kind, source, cfg)
}
