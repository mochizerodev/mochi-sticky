package tui

import (
	"testing"

	"mochi-sticky/internal/board"

	tea "github.com/charmbracelet/bubbletea"
)

func TestClampIndex(t *testing.T) {
	if got := clampIndex(-1, 3); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
	if got := clampIndex(5, 3); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
	if got := clampIndex(1, 3); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestNormalizedKeyArrows(t *testing.T) {
	if got := normalizedKey(tea.KeyMsg{Type: tea.KeyUp}); got != "k" {
		t.Fatalf("expected k, got %s", got)
	}
	if got := normalizedKey(tea.KeyMsg{Type: tea.KeyDown}); got != "j" {
		t.Fatalf("expected j, got %s", got)
	}
	if got := normalizedKey(tea.KeyMsg{Type: tea.KeyLeft}); got != "h" {
		t.Fatalf("expected h, got %s", got)
	}
	if got := normalizedKey(tea.KeyMsg{Type: tea.KeyRight}); got != "l" {
		t.Fatalf("expected l, got %s", got)
	}
}

func TestClampSelectionEmpty(t *testing.T) {
	col := columnModel{}
	clampSelection(&col)
	if col.Selected != 0 {
		t.Fatalf("expected selection 0, got %d", col.Selected)
	}
}

func TestMoveSelectionBoundaries(t *testing.T) {
	m := Model{
		columns: []columnModel{
			{Tasks: []board.Task{{ID: "T-1"}, {ID: "T-2"}}, Selected: 0},
		},
	}
	m = m.moveSelection(-1)
	if m.columns[0].Selected != 0 {
		t.Fatalf("expected clamp at 0, got %d", m.columns[0].Selected)
	}
	m = m.moveSelection(5)
	if m.columns[0].Selected != 1 {
		t.Fatalf("expected clamp at last, got %d", m.columns[0].Selected)
	}
}

func TestNormalizeKeyUppercase(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}}
	if got := normalizedKey(msg); got != "j" {
		t.Fatalf("expected lowercase j, got %s", got)
	}
}

func TestRestoreSelection(t *testing.T) {
	m := Model{
		columns: []columnModel{
			{Tasks: []board.Task{{ID: "T-1"}, {ID: "T-2"}}},
			{Tasks: []board.Task{{ID: "T-3"}}},
		},
		selectedTaskID: "T-3",
	}
	m.restoreSelection()
	if m.active != 1 || m.columns[1].Selected != 0 {
		t.Fatalf("expected selection restored to column 1 index 0, got %d/%d", m.active, m.columns[1].Selected)
	}
	if m.selectedTaskID != "" {
		t.Fatalf("expected selectedTaskID cleared")
	}
}

func TestHandleStatusPickerKeyEnterUpdatesStatus(t *testing.T) {
	m := Model{
		screen: screenStatusPicker,
		columns: []columnModel{
			{Key: "todo", Tasks: []board.Task{{ID: "T-1", Status: "todo"}}},
			{Key: "done", Tasks: []board.Task{}},
		},
		active:      0,
		statusIndex: 1,
	}
	_, cmd := m.handleStatusPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command on enter")
	}
}

func TestHandleTaskCreateKeyEmptyTitle(t *testing.T) {
	m := Model{
		screen:    screenTaskCreate,
		taskTitle: "   ",
	}
	_, cmd := m.handleTaskCreateKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("expected no command when title is empty")
	}
}

func TestHandleArchiveKeyRestore(t *testing.T) {
	m := Model{
		screen:       screenArchive,
		archived:     []board.Task{{ID: "T-1"}},
		archiveIndex: 0,
	}
	_, cmd := m.handleArchiveKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command when restoring archived task")
	}
}

func TestHandleTaskActionsKeyEnter(t *testing.T) {
	m := Model{
		screen: screenTaskActions,
		columns: []columnModel{
			{Tasks: []board.Task{{ID: "T-1", Status: "todo"}}},
			{Tasks: []board.Task{}},
		},
		active:     0,
		taskAction: 0,
	}
	_, cmd := m.handleTaskActionsKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command for task action")
	}
}
