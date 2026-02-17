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

func TestSwitchToTabPreservesBoardScreen(t *testing.T) {
	m := Model{
		screen:         screenTaskDetail,
		activeTab:      tabBoards,
		boardTabScreen: screenBoard,
		wikiTabScreen:  screenWiki,
		wikiLoaded:     true,
		wikiIndex:      2,
		wikiItems:      []wikiNavItem{{Kind: wikiItemSection, Title: "Docs"}},
		confirmAction:  confirmNone,
	}

	next, cmd := m.switchToTab(tabWiki)
	if cmd != nil {
		t.Fatalf("expected no load command when wiki data is already loaded")
	}
	wikiModel := next.(Model)
	if wikiModel.screen != screenWiki {
		t.Fatalf("expected wiki screen, got %v", wikiModel.screen)
	}
	if wikiModel.wikiIndex != 2 {
		t.Fatalf("expected wiki selection preserved, got %d", wikiModel.wikiIndex)
	}

	back, cmd := wikiModel.switchToTab(tabBoards)
	if cmd != nil {
		t.Fatalf("expected no command when returning to boards")
	}
	boardModel := back.(Model)
	if boardModel.screen != screenTaskDetail {
		t.Fatalf("expected board screen restored, got %v", boardModel.screen)
	}
}

func TestSwitchToTabLoadsWikiWhenNeeded(t *testing.T) {
	m := Model{
		screen:    screenBoard,
		activeTab: tabBoards,
	}
	next, cmd := m.switchToTab(tabWiki)
	if cmd == nil {
		t.Fatalf("expected load command for first wiki switch")
	}
	wikiModel := next.(Model)
	if !wikiModel.loading {
		t.Fatalf("expected loading state while fetching wiki")
	}
	if wikiModel.screen != screenWiki {
		t.Fatalf("expected wiki root screen while loading, got %v", wikiModel.screen)
	}
}

func TestTabFromKeyDigitOutsideTextEntry(t *testing.T) {
	target, ok := tabFromKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}, screenBoard)
	if !ok {
		t.Fatalf("expected digit key to map to a tab")
	}
	if target != tabWiki {
		t.Fatalf("expected wiki tab, got %v", target)
	}
}

func TestTabFromKeyDigitInsideTextEntry(t *testing.T) {
	if _, ok := tabFromKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}, screenTaskCreate); ok {
		t.Fatalf("expected no tab mapping in text-entry screens")
	}
}
