package main

import (
	"fmt"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/hatolife/HatoNote/internal/markdownquality"
	"github.com/hatolife/HatoNote/internal/settings"
)

func markdownFormatOptions(cfg settings.Settings) markdownquality.FormatOptions {
	return markdownquality.FormatOptions{
		TrimTrailingWhitespace: cfg.FormatTrimTrailingWhitespace,
		MaxBlankLines:          cfg.FormatMaxBlankLines,
		FinalNewline:           cfg.FormatFinalNewline,
	}
}

func markdownLintOptions(cfg settings.Settings) markdownquality.LintOptions {
	return markdownquality.LintOptions{
		TrailingWhitespace: cfg.LintTrailingWhitespace,
		LongLines:          cfg.LintLongLines,
		MaxLineLength:      cfg.LintMaxLineLength,
		HeadingStep:        cfg.LintHeadingStep,
		FinalNewline:       cfg.LintFinalNewline,
	}
}

func formatMarkdownChecked(text string, cfg settings.Settings) (string, error) {
	if len(text) > bundle.MaxFile {
		return "", fmt.Errorf("本文が大きすぎます")
	}
	formatted := markdownquality.Format(text, markdownFormatOptions(cfg))
	if len(formatted) > bundle.MaxFile {
		return "", fmt.Errorf("整形後の本文が大きすぎます")
	}
	return formatted, nil
}

// FormatMarkdown は内蔵エディターから渡されたMarkdownを現在設定で整形します。
func (a *App) FormatMarkdown(text string) (string, error) {
	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()
	return formatMarkdownChecked(text, cfg)
}

// FormatNativeMarkdown は現在のNeovimバッファをUndo可能な変更として整形します。
func (a *App) FormatNativeMarkdown() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.native == nil {
		return fmt.Errorf("Neovimを開始してください")
	}
	current, err := a.native.Current()
	if err != nil {
		return err
	}
	formatted, err := formatMarkdownChecked(current.Text, a.cfg)
	if err != nil {
		return err
	}
	if formatted == current.Text {
		return nil
	}
	return a.native.ReplaceCurrent(formatted)
}

// LintMarkdown は現在設定でMarkdownを検査します。
func (a *App) LintMarkdown(text string) ([]markdownquality.Diagnostic, error) {
	if len(text) > bundle.MaxFile {
		return nil, fmt.Errorf("本文が大きすぎます")
	}
	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()
	return markdownquality.Lint(text, markdownLintOptions(cfg)), nil
}

// LintNativeMarkdown は現在のNeovimバッファを検査します。
func (a *App) LintNativeMarkdown() ([]markdownquality.Diagnostic, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.native == nil {
		return nil, fmt.Errorf("Neovimを開始してください")
	}
	current, err := a.native.Current()
	if err != nil {
		return nil, err
	}
	return markdownquality.Lint(current.Text, markdownLintOptions(a.cfg)), nil
}
