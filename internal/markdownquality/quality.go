package markdownquality

import (
	"strings"
	"unicode/utf8"
)

type FormatOptions struct {
	TrimTrailingWhitespace bool
	MaxBlankLines          int
	FinalNewline           bool
}

type LintOptions struct {
	TrailingWhitespace bool
	LongLines          bool
	MaxLineLength      int
	HeadingStep        bool
	FinalNewline       bool
}

type Diagnostic struct {
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type fenceState struct {
	active bool
	marker byte
	length int
}

func fence(line string, state fenceState) (fenceState, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || trimmed == "" {
		return state, false
	}
	marker := trimmed[0]
	if marker != '`' && marker != '~' {
		return state, false
	}
	count := 0
	for count < len(trimmed) && trimmed[count] == marker {
		count++
	}
	if count < 3 {
		return state, false
	}
	if !state.active {
		return fenceState{active:true, marker:marker, length:count}, true
	}
	if marker == state.marker && count >= state.length && strings.TrimSpace(trimmed[count:]) == "" {
		return fenceState{}, true
	}
	return state, false
}

func Format(text string, options FormatOptions) string {
	hadFinalNewline := strings.HasSuffix(text, "\n")
	lines := strings.Split(text, "\n")
	if hadFinalNewline && len(lines) > 0 {
		lines = lines[:len(lines)-1]
	}
	result := make([]string, 0, len(lines)+1)
	state := fenceState{}
	blankLines := 0
	for _, original := range lines {
		nextState, fenceLine := fence(original, state)
		line := original
		if !state.active && !fenceLine && options.TrimTrailingWhitespace {
			line = strings.TrimRight(line, " \t\r")
		}
		if !state.active && !fenceLine && strings.TrimSpace(line) == "" {
			if blankLines >= options.MaxBlankLines {
				state = nextState
				continue
			}
			blankLines++
		} else {
			blankLines = 0
		}
		result = append(result, line)
		state = nextState
	}
	formatted := strings.Join(result, "\n")
	if options.FinalNewline {
		return strings.TrimRight(formatted, "\n") + "\n"
	}
	if hadFinalNewline {
		return formatted + "\n"
	}
	return formatted
}

func headingLevel(line string) int {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || len(trimmed) < 2 || trimmed[0] != '#' {
		return 0
	}
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level < 1 || level > 6 || level >= len(trimmed) || trimmed[level] != ' ' {
		return 0
	}
	return level
}

func Lint(text string, options LintOptions) []Diagnostic {
	lines := strings.Split(text, "\n")
	result := []Diagnostic{}
	state := fenceState{}
	lastHeading := 0
	for index, original := range lines {
		if index == len(lines)-1 && original == "" && strings.HasSuffix(text, "\n") {
			break
		}
		nextState, fenceLine := fence(original, state)
		if !state.active && !fenceLine {
			if options.TrailingWhitespace && strings.TrimRight(original, " \t") != original {
				result = append(result, Diagnostic{Line:index+1, Rule:"trailing-whitespace", Message:"行末に空白があります"})
			}
			if options.LongLines && utf8.RuneCountInString(original) > options.MaxLineLength {
				result = append(result, Diagnostic{Line:index+1, Rule:"line-length", Message:"行が設定した最大文字数を超えています"})
			}
			if options.HeadingStep {
				level := headingLevel(original)
				if level > 0 {
					if lastHeading > 0 && level > lastHeading+1 {
						result = append(result, Diagnostic{Line:index+1, Rule:"heading-step", Message:"見出しレベルが途中の階層を飛ばしています"})
					}
					lastHeading = level
				}
			}
		}
		state = nextState
	}
	if options.FinalNewline && text != "" && !strings.HasSuffix(text, "\n") {
		result = append(result, Diagnostic{Line:len(lines), Rule:"final-newline", Message:"ファイル末尾に改行がありません"})
	}
	return result
}
