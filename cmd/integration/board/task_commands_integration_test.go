package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mochi-sticky/internal/board"
)

func TestTaskListCommandFilters(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	taskAlpha := createTask(t, repoRoot, storageRoot, "Alpha task", []string{"backend", "api"}, 2)
	taskBeta := createTask(t, repoRoot, storageRoot, "Beta task", []string{"frontend"}, 1)
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "move", taskBeta, "doing"); err != nil {
		t.Fatalf("task move: %v", err)
	}

	cases := []struct {
		name        string
		args        []string
		want        []string
		notContains []string
	}{
		{
			name: "status filter",
			args: []string{"task", "list", "--status", "doing"},
			want: []string{taskBeta},
			notContains: []string{
				taskAlpha,
				"No tasks found.",
			},
		},
		{
			name: "title filter",
			args: []string{"task", "list", "--title", "Alpha"},
			want: []string{taskAlpha},
			notContains: []string{
				taskBeta,
				"No tasks found.",
			},
		},
		{
			name: "tag any filter",
			args: []string{"task", "list", "--tag", "backend", "--tag-mode", "any"},
			want: []string{taskAlpha},
			notContains: []string{
				"No tasks found.",
			},
		},
		{
			name: "tag all filter",
			args: []string{"task", "list", "--tag", "backend", "--tag", "api", "--tag-mode", "all"},
			want: []string{taskAlpha},
			notContains: []string{
				taskBeta,
				"No tasks found.",
			},
		},
		{
			name: "date range filter",
			args: []string{"task", "list", "--from", "1970-01-01", "--to", "2100-01-01"},
			want: []string{taskAlpha, taskBeta},
			notContains: []string{
				"No tasks found.",
			},
		},
		{
			name: "sort desc",
			args: []string{"task", "list", "--sort", "title", "--desc"},
			want: []string{taskAlpha, taskBeta},
			notContains: []string{
				"No tasks found.",
			},
		},
	}

	// Act + Assert
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runMochiSticky(t, repoRoot, storageRoot, tc.args...)
			if err != nil {
				t.Fatalf("task list: %v", err)
			}
			clean := stripANSI(out)
			for _, want := range tc.want {
				if !strings.Contains(clean, want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, clean)
				}
			}
			for _, nope := range tc.notContains {
				if strings.Contains(clean, nope) {
					t.Fatalf("expected output to not contain %q, got:\n%s", nope, clean)
				}
			}
		})
	}
}

func TestTaskShowCommandOutputsDetails(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	taskID := createTask(t, repoRoot, storageRoot, "Show task", nil, 2)

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "task", "show", taskID)
	if err != nil {
		t.Fatalf("task show: %v", err)
	}

	// Assert
	if !strings.Contains(out, "Title: Show task") {
		t.Fatalf("expected title in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Status: todo") {
		t.Fatalf("expected status in output, got:\n%s", out)
	}
}

func TestTaskMoveCommandUpdatesStatus(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	taskID := createTask(t, repoRoot, storageRoot, "Move task", nil, 0)

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "task", "move", taskID, "doing")
	if err != nil {
		t.Fatalf("task move: %v", err)
	}

	// Assert
	if !strings.Contains(out, fmt.Sprintf("Moved task %s to doing", taskID)) {
		t.Fatalf("expected move confirmation, got:\n%s", out)
	}
	task := readTask(t, storageRoot, taskID)
	if task.Status != "doing" {
		t.Fatalf("expected status %q, got %q", "doing", task.Status)
	}
}

func TestTaskDepsCommandSetAndShow(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	depID := createTask(t, repoRoot, storageRoot, "Dep task", nil, 0)
	taskID := createTask(t, repoRoot, storageRoot, "Main task", nil, 0)

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "deps", taskID, "--set", depID); err != nil {
		t.Fatalf("task deps set: %v", err)
	}
	out, err := runMochiSticky(t, repoRoot, storageRoot, "task", "deps", taskID)
	if err != nil {
		t.Fatalf("task deps show: %v", err)
	}

	// Assert
	if !strings.Contains(out, fmt.Sprintf("Task: %s", taskID)) {
		t.Fatalf("expected task ID in output, got:\n%s", out)
	}
	if !strings.Contains(out, fmt.Sprintf("Depends on: %s", depID)) {
		t.Fatalf("expected dependency in output, got:\n%s", out)
	}
}

func TestTaskReadyCommandListsSatisfiedDeps(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	doneID := createTask(t, repoRoot, storageRoot, "Ready dependency", nil, 0)
	blockedID := createTask(t, repoRoot, storageRoot, "Ready target", nil, 0)
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "deps", blockedID, "--set", doneID); err != nil {
		t.Fatalf("task deps set: %v", err)
	}
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "move", doneID, "done"); err != nil {
		t.Fatalf("task move: %v", err)
	}

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "task", "ready")
	if err != nil {
		t.Fatalf("task ready: %v", err)
	}

	// Assert
	if !strings.Contains(out, blockedID) {
		t.Fatalf("expected ready task in output, got:\n%s", out)
	}
}

func TestTaskStatusesCommandListsDefaults(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "task", "statuses")
	if err != nil {
		t.Fatalf("task statuses: %v", err)
	}

	// Assert
	for _, status := range []string{"todo", "doing", "done"} {
		if !strings.Contains(out, status) {
			t.Fatalf("expected status %q in output, got:\n%s", status, out)
		}
	}
}

func TestTaskPriorityCommandUpdatesValue(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	taskID := createTask(t, repoRoot, storageRoot, "Priority task", nil, 1)

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "task", "priority", taskID, "3")
	if err != nil {
		t.Fatalf("task priority: %v", err)
	}

	// Assert
	if !strings.Contains(out, fmt.Sprintf("Updated priority for %s to 3", taskID)) {
		t.Fatalf("expected priority confirmation, got:\n%s", out)
	}
	task := readTask(t, storageRoot, taskID)
	if task.Priority != 3 {
		t.Fatalf("expected priority 3, got %d", task.Priority)
	}
}

func TestTaskDeleteCommandRemovesFile(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	taskID := createTask(t, repoRoot, storageRoot, "Delete task", nil, 0)

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "delete", taskID, "--force"); err != nil {
		t.Fatalf("task delete: %v", err)
	}

	// Assert
	tasksDir := filepath.Join(storageRoot, "boards", "default", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		t.Fatalf("read tasks dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 task files, got %d", len(entries))
	}
}

func TestTaskAddCommandCreatesTaskFile(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "add", "Title"); err != nil {
		t.Fatalf("task add: %v", err)
	}

	// Assert
	tasksDir := filepath.Join(storageRoot, "boards", "default", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		t.Fatalf("read tasks dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 task file, got %d", len(entries))
	}

	taskPath := filepath.Join(tasksDir, entries[0].Name())
	data, err := os.ReadFile(taskPath)
	if err != nil {
		t.Fatalf("read task file: %v", err)
	}
	parsed, err := (&board.Parser{}).Parse(data)
	if err != nil {
		t.Fatalf("parse task file: %v", err)
	}
	if parsed.Title != "Title" {
		t.Fatalf("expected title %q, got %q", "Title", parsed.Title)
	}
	if parsed.Status != board.DefaultStatus {
		t.Fatalf("expected status %q, got %q", board.DefaultStatus, parsed.Status)
	}
	if parsed.Priority != board.DefaultPriority {
		t.Fatalf("expected priority %d, got %d", board.DefaultPriority, parsed.Priority)
	}
	if parsed.ID == "" {
		t.Fatalf("expected task ID to be set")
	}
}

func TestTaskArchiveCommandsFlow(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	taskID := createTask(t, repoRoot, storageRoot, "Archive task", nil, 0)

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "archive", "task", taskID, "--force"); err != nil {
		t.Fatalf("archive task: %v", err)
	}
	listOut, err := runMochiSticky(t, repoRoot, storageRoot, "task", "archive", "list")
	if err != nil {
		t.Fatalf("archive list: %v", err)
	}

	// Assert
	cleanList := stripANSI(listOut)
	if !strings.Contains(cleanList, taskID) {
		t.Fatalf("expected archived task in list, got:\n%s", cleanList)
	}

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "archive", "restore", taskID, "--force"); err != nil {
		t.Fatalf("archive restore: %v", err)
	}
	activeOut, err := runMochiSticky(t, repoRoot, storageRoot, "task", "list")
	if err != nil {
		t.Fatalf("task list: %v", err)
	}

	// Assert
	activeClean := stripANSI(activeOut)
	if !strings.Contains(activeClean, taskID) {
		t.Fatalf("expected restored task in list, got:\n%s", activeClean)
	}

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "archive", "before", "2100-01-01", "--force"); err != nil {
		t.Fatalf("archive before: %v", err)
	}
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "task", "archive", "delete", taskID, "--force"); err != nil {
		t.Fatalf("archive delete: %v", err)
	}
	afterDeleteOut, err := runMochiSticky(t, repoRoot, storageRoot, "task", "archive", "list")
	if err != nil {
		t.Fatalf("archive list: %v", err)
	}

	// Assert
	if !strings.Contains(afterDeleteOut, "No archived tasks found.") {
		t.Fatalf("expected empty archive message, got:\n%s", afterDeleteOut)
	}
}
