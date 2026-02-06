package tui

import (
	"testing"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/board"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleConfirmCancelTask(t *testing.T) {
	m := Model{
		screen:        screenConfirm,
		confirmAction: confirmDeleteTask,
	}
	m2, _ := m.handleConfirmKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m2.(Model).screen != screenBoard {
		t.Fatalf("expected screenBoard after cancel")
	}
}

func TestADRDeleteActionPromptsConfirm(t *testing.T) {
	items := adrActionItems()
	deleteIndex := -1
	for i, item := range items {
		if item == "delete adr" {
			deleteIndex = i
			break
		}
	}
	if deleteIndex == -1 {
		t.Fatalf("delete adr action not found")
	}
	m := Model{
		screen:    screenADRActions,
		adrAction: deleteIndex,
		adrColumns: []adrColumnModel{
			{
				Key:   "proposed",
				Title: "Proposed",
				ADRs:  []adr.ADR{{ID: 12, Title: "Test", Status: "proposed"}},
			},
		},
	}
	model, _ := m.handleADRActionSelection()
	got := model.(Model)
	if got.screen != screenConfirm {
		t.Fatalf("expected screenConfirm, got %v", got.screen)
	}
	if got.confirmAction != confirmDeleteADR {
		t.Fatalf("expected confirmDeleteADR, got %v", got.confirmAction)
	}
	if got.confirmADR != 12 {
		t.Fatalf("expected confirm ADR id 12, got %d", got.confirmADR)
	}
}

func TestHandleTaskDetailKeyX(t *testing.T) {
	m := Model{
		screen: screenTaskDetail,
	}
	m2, _ := m.handleTaskDetailKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m2.(Model).screen != screenTaskActions {
		t.Fatalf("expected screenTaskActions")
	}
}

func TestApplyStatusUpdateMissingTaskNoChange(t *testing.T) {
	cols := []board.Column{{Key: "todo", Title: "Todo"}}
	m := Model{columns: buildColumns(cols, []board.Task{{ID: "T-1", Status: "todo"}})}
	m2 := m.applyStatusUpdate("missing", "done")
	if len(m2.columns[0].Tasks) != 1 {
		t.Fatalf("expected unchanged tasks")
	}
}
