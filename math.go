package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hatolife/HatoNote/internal/settings"
)

const maxMathSource = 256 << 10
const maxMathML = 2 << 20

var mathCache = struct {
	sync.Mutex
	values map[string]string
}{values: map[string]string{}}

func mathCacheKey(source string, display bool, cfg settings.Settings) string {
	value := strings.Join([]string{source, fmt.Sprint(display), cfg.KatexPath}, "\x00")
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func mathCacheGet(key string) (string, bool) {
	mathCache.Lock()
	defer mathCache.Unlock()
	value, ok := mathCache.values[key]
	return value, ok
}

func mathCachePut(key, value string) {
	mathCache.Lock()
	defer mathCache.Unlock()
	if len(mathCache.values) >= 128 {
		for existing := range mathCache.values {
			delete(mathCache.values, existing)
			break
		}
	}
	mathCache.values[key] = value
}

type limitedBuffer struct {
	bytes.Buffer
	max int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > b.max {
		return 0, fmt.Errorf("出力サイズが上限を超えています")
	}
	return b.Buffer.Write(p)
}

func renderMath(source string, display bool, cfg settings.Settings) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", fmt.Errorf("数式が空です")
	}
	if len(source) > maxMathSource {
		return "", fmt.Errorf("数式は256KiB以下にしてください")
	}
	key := mathCacheKey(source, display, cfg)
	if value, ok := mathCacheGet(key); ok {
		return value, nil
	}
	executable := resolveDiagramExecutable(cfg.KatexPath, "katex")
	if executable == "" {
		return "", fmt.Errorf("KaTeX CLIが見つかりません。設定からkatexを指定してください")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := []string{"-F", "mathml", "-S", "-s", "100", "-e", "1000"}
	if display {
		args = append(args, "-d")
	}
	command := diagramCommandContext(ctx, executable, args...)
	command.Stdin = strings.NewReader(source)
	stdout := &limitedBuffer{max:maxMathML}
	stderr := &limitedBuffer{max:64 << 10}
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("数式描画が20秒でタイムアウトしました")
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("KaTeXの描画に失敗しました: %s", message)
	}
	value := strings.TrimSpace(stdout.String())
	if !strings.Contains(value, "<math") {
		return "", fmt.Errorf("KaTeXがMathMLを生成しませんでした")
	}
	mathCachePut(key, value)
	return value, nil
}

// RenderMath はTeX数式をMathMLへ変換します。
func (a *App) RenderMath(source string, display bool) (string, error) {
	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()
	return renderMath(source, display, cfg)
}
