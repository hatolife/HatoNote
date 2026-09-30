package main

import "testing"

func TestHistoryLineDiffBasicChanges(t *testing.T) {
	got := historyLineDiff("a\nb\nc\n", "a\nB\nc\n")
	if len(got) != 4 {
		t.Fatalf("diff length = %d: %+v", len(got), got)
	}
	if got[0].Status != "equal" || got[1].Status != "modified" || got[2].Status != "equal" {
		t.Fatalf("diff = %+v", got)
	}
	if got[1].HistoryLine != 2 || got[1].CurrentLine != 2 || got[1].History != "b" || got[1].Current != "B" {
		t.Fatalf("modified row = %+v", got[1])
	}
}

func TestHistoryLineDiffAddDeleteAndResync(t *testing.T) {
	added := historyLineDiff("a\nc", "a\nb\nc")
	if len(added) != 3 || added[1].Status != "added" || added[1].Current != "b" {
		t.Fatalf("added = %+v", added)
	}
	deleted := historyLineDiff("a\nb\nc", "a\nc")
	if len(deleted) != 3 || deleted[1].Status != "deleted" || deleted[1].History != "b" {
		t.Fatalf("deleted = %+v", deleted)
	}
	multiple := historyLineDiff(
		"one\ntwo\nthree\nfour\nfive",
		"one\nTWO\nthree\ninserted\nfour\nFIVE",
	)
	statuses := make([]string, len(multiple))
	for i, row := range multiple {
		statuses[i] = row.Status
	}
	want := []string{"equal", "modified", "equal", "added", "equal", "modified"}
	if len(statuses) != len(want) {
		t.Fatalf("statuses = %v", statuses)
	}
	for i := range want {
		if statuses[i] != want[i] {
			t.Fatalf("statuses = %v, want %v", statuses, want)
		}
	}
}

func TestHistoryLineDiffLargeInputUsesBoundedResync(t *testing.T) {
	history := ""
	current := ""
	for i := 0; i < 5000; i++ {
		line := "line"
		if i%7 == 0 {
			line += "-anchor"
		}
		history += line + "\n"
		current += line + "\n"
	}
	got := historyLineDiff(history, current)
	if len(got) == 0 {
		t.Fatal("empty diff")
	}
	for _, row := range got {
		if row.Status != "equal" {
			t.Fatalf("unexpected change: %+v", row)
		}
	}
}
