package tui

import (
	"context"
	"errors"
	"testing"

	"mochi-sticky/internal/board"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBuildColumnsAddsUnknown(t *testing.T) {
	columns := []board.Column{
		{Key: "todo", Title: "Todo"},
	}
	tasks := []board.Task{
		{ID: "T-1", Title: "Known", Status: "todo"},
		{ID: "T-2", Title: "Unknown", Status: "mystery"},
	}

	result := buildColumns(columns, tasks)
	if len(result) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result))
	}
	if result[1].Key != "unknown" {
		t.Fatalf("expected unknown column, got %s", result[1].Key)
	}
	if len(result[1].Tasks) != 1 || result[1].Tasks[0].ID != "T-2" {
		t.Fatalf("expected unknown task in unknown column")
	}
}

func TestBuildColumnsSortsReadyFirst(t *testing.T) {
	columns := []board.Column{
		{Key: "todo", Title: "Todo"},
	}
	tasks := []board.Task{
		{ID: "T-1", Title: "Blocked", Status: "todo", DependsOn: []string{"T-2"}},
		{ID: "T-2", Title: "Ready", Status: "todo"},
	}

	result := buildColumns(columns, tasks)
	if len(result) != 1 {
		t.Fatalf("expected 1 column, got %d", len(result))
	}
	if len(result[0].Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result[0].Tasks))
	}
	if result[0].Tasks[0].ID != "T-2" {
		t.Fatalf("expected ready task first, got %s", result[0].Tasks[0].ID)
	}
}

func TestApplyStatusUpdateMovesTask(t *testing.T) {
	cols := []board.Column{
		{Key: "todo", Title: "Todo"},
		{Key: "done", Title: "Done"},
	}
	tasks := []board.Task{
		{ID: "T-1", Title: "Task", Status: "todo"},
	}

	m := Model{columns: buildColumns(cols, tasks)}
	m = m.applyStatusUpdate("T-1", "done")

	if len(m.columns[0].Tasks) != 0 {
		t.Fatalf("expected task removed from todo")
	}
	if len(m.columns[1].Tasks) != 1 || m.columns[1].Tasks[0].ID != "T-1" {
		t.Fatalf("expected task in done column")
	}
}

func TestEffectivePriorityDefaults(t *testing.T) {
	if got := effectivePriority(0); got != board.DefaultPriority {
		t.Fatalf("expected default priority, got %d", got)
	}
	if got := effectivePriority(2); got != 2 {
		t.Fatalf("expected priority 2, got %d", got)
	}
}

func TestBuildColumnsSortsByPriority(t *testing.T) {
	columns := []board.Column{
		{Key: "todo", Title: "Todo"},
	}
	tasks := []board.Task{
		{ID: "T-1", Title: "Low", Status: "todo", Priority: 3},
		{ID: "T-2", Title: "High", Status: "todo", Priority: 1},
	}
	result := buildColumns(columns, tasks)
	if result[0].Tasks[0].ID != "T-2" {
		t.Fatalf("expected higher priority first, got %s", result[0].Tasks[0].ID)
	}
}

func TestApplyStatusUpdateUnknownStatusNoChange(t *testing.T) {
	cols := []board.Column{
		{Key: "todo", Title: "Todo"},
	}
	tasks := []board.Task{
		{ID: "T-1", Title: "Task", Status: "todo"},
	}
	m := Model{columns: buildColumns(cols, tasks)}
	m2 := m.applyStatusUpdate("T-1", "missing")
	if len(m2.columns) != 1 || len(m2.columns[0].Tasks) != 0 {
		t.Fatalf("expected task removed when status missing")
	}
}

// TestContextCanceledErrorIsIgnored verifies that context.Canceled errors
// from in-flight operations (e.g., during refresh) are silently ignored
// and don't set the model's error field.
func TestContextCanceledErrorIsIgnored(t *testing.T) {
	m := Model{}

	// Simulate receiving a context.Canceled error
	msg := errMsg{err: context.Canceled}
	result, cmd := m.Update(msg)

	resultModel := result.(Model)
	if resultModel.err != nil {
		t.Fatalf("expected no error after context.Canceled, got: %v", resultModel.err)
	}
	if resultModel.loading {
		t.Fatalf("expected loading to be false after context.Canceled")
	}
	if cmd != nil {
		t.Fatalf("expected no command after context.Canceled")
	}
}

// TestOtherErrorsArePreserved verifies that non-Canceled errors are still
// properly set in the model's error field.
func TestOtherErrorsArePreserved(t *testing.T) {
	m := Model{}

	expectedErr := errors.New("some other error")
	msg := errMsg{err: expectedErr}
	result, cmd := m.Update(msg)

	resultModel := result.(Model)
	if resultModel.err != expectedErr {
		t.Fatalf("expected error to be preserved, got: %v", resultModel.err)
	}
	if resultModel.loading {
		t.Fatalf("expected loading to be false after error")
	}
	if cmd != nil {
		t.Fatalf("expected no command after error")
	}
}

// TestRefreshCancelsInFlightOperations verifies that pressing ctrl+r or F5
// properly cancels any in-flight operations before starting new ones.
func TestRefreshCancelsInFlightOperations(t *testing.T) {
	m := Model{
		repo:      &board.Repository{},
		boardRepo: &board.BoardRepository{},
	}

	// Set up an in-flight operation
	ctx, cancel := context.WithCancel(context.Background())
	m.inFlightCancel = cancel

	// Simulate refresh key press
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	result, _ := m.Update(msg)

	resultModel := result.(Model)

	// Verify the old context was cancelled (inFlightCancel should be replaced)
	select {
	case <-ctx.Done():
		// Expected: context should be cancelled
	default:
		t.Fatalf("expected old context to be cancelled")
	}

	// Verify a new in-flight operation was started
	if resultModel.inFlightCancel == nil {
		t.Fatalf("expected new in-flight operation to be started")
	}
	if !resultModel.loading {
		t.Fatalf("expected loading to be true during refresh")
	}
}
