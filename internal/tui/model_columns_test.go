package tui

import (
	"testing"

	"mochi-sticky/internal/board"
)

func TestFindColumnAndFindTask(t *testing.T) {
	cols := []board.Column{
		{Key: "todo", Title: "Todo"},
		{Key: "done", Title: "Done"},
	}
	tasks := []board.Task{
		{ID: "T-1", Title: "Task", Status: "todo"},
	}
	modelCols := buildColumns(cols, tasks)
	colIdx := findColumn(modelCols, "done")
	if colIdx != 1 {
		t.Fatalf("expected done column index 1, got %d", colIdx)
	}
	foundCol, foundIdx, task := findTask(modelCols, "T-1")
	if foundCol != 0 || foundIdx != 0 || task.ID != "T-1" {
		t.Fatalf("expected to find task in first column")
	}
}

func TestBuildTaskIndex(t *testing.T) {
	cols := []columnModel{
		{Tasks: []board.Task{{ID: "T-1"}, {ID: "T-2"}}},
	}
	index := buildTaskIndex(cols)
	if len(index) != 2 {
		t.Fatalf("expected 2 tasks in index, got %d", len(index))
	}
	if _, ok := index["T-2"]; !ok {
		t.Fatalf("expected T-2 in index")
	}
}
