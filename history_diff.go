package main

import "strings"

const historyDiffLookahead = 80

type HistoryLineDiff struct {
	Status      string `json:"status"`
	HistoryLine int    `json:"historyLine,omitempty"`
	CurrentLine int    `json:"currentLine,omitempty"`
	History     string `json:"history"`
	Current     string `json:"current"`
}

func historyDiffLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

func appendHistoryLine(result []HistoryLineDiff, status string, historyLine, currentLine int, history, current string) []HistoryLineDiff {
	return append(result, HistoryLineDiff{
		Status: status, HistoryLine: historyLine, CurrentLine: currentLine,
		History: history, Current: current,
	})
}

func findHistoryResync(history []string, hi int, current []string, ci int) (int, int, bool) {
	hEnd := min(len(history), hi+historyDiffLookahead+1)
	cEnd := min(len(current), ci+historyDiffLookahead+1)
	currentOffset := make(map[string]int, cEnd-ci)
	for index := ci; index < cEnd; index++ {
		line := current[index]
		if _, exists := currentOffset[line]; !exists {
			currentOffset[line] = index - ci
		}
	}
	bestH, bestC, bestScore := -1, -1, int(^uint(0)>>1)
	for index := hi; index < hEnd; index++ {
		offset, exists := currentOffset[history[index]]
		if !exists {
			continue
		}
		hOffset := index - hi
		if hOffset == 0 && offset == 0 {
			continue
		}
		score := hOffset + offset
		if score < bestScore {
			bestH, bestC, bestScore = hOffset, offset, score
		}
	}
	return bestH, bestC, bestH >= 0
}

// historyLineDiff は最大80行の先読みだけで再同期し、二次メモリを使わず行差分を生成します。
func historyLineDiff(historyText, currentText string) []HistoryLineDiff {
	history := historyDiffLines(historyText)
	current := historyDiffLines(currentText)
	result := make([]HistoryLineDiff, 0, max(len(history), len(current)))
	hi, ci := 0, 0
	for hi < len(history) || ci < len(current) {
		if hi < len(history) && ci < len(current) && history[hi] == current[ci] {
			result = appendHistoryLine(result, "equal", hi+1, ci+1, history[hi], current[ci])
			hi++
			ci++
			continue
		}
		if hi >= len(history) {
			result = appendHistoryLine(result, "added", 0, ci+1, "", current[ci])
			ci++
			continue
		}
		if ci >= len(current) {
			result = appendHistoryLine(result, "deleted", hi+1, 0, history[hi], "")
			hi++
			continue
		}

		hAhead, cAhead, found := findHistoryResync(history, hi, current, ci)
		if !found {
			result = appendHistoryLine(result, "modified", hi+1, ci+1, history[hi], current[ci])
			hi++
			ci++
			continue
		}

		paired := min(hAhead, cAhead)
		for offset := 0; offset < paired; offset++ {
			result = appendHistoryLine(result, "modified", hi+offset+1, ci+offset+1, history[hi+offset], current[ci+offset])
		}
		for offset := paired; offset < hAhead; offset++ {
			result = appendHistoryLine(result, "deleted", hi+offset+1, 0, history[hi+offset], "")
		}
		for offset := paired; offset < cAhead; offset++ {
			result = appendHistoryLine(result, "added", 0, ci+offset+1, "", current[ci+offset])
		}
		hi += hAhead
		ci += cAhead
	}
	return result
}
