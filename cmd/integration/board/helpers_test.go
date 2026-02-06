package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/testutil"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func setupStorage(t *testing.T) (repoRoot, storageRoot string) {
	return testutil.SetupStorage(t)
}

func runMochiSticky(t *testing.T, repoRoot, storageRoot string, args ...string) (string, error) {
	return testutil.RunMochiSticky(t, repoRoot, storageRoot, args...)
}

func createTask(t *testing.T, repoRoot, storageRoot, title string, tags []string, priority int) string {
	t.Helper()

	args := []string{"task", "add", title}
	if len(tags) > 0 {
		args = append(args, "--tags", strings.Join(tags, ","))
	}
	if priority > 0 {
		args = append(args, "--priority", fmt.Sprintf("%d", priority))
	}
	out, err := runMochiSticky(t, repoRoot, storageRoot, args...)
	if err != nil {
		t.Fatalf("task add: %v", err)
	}
	return parseCreatedTaskID(t, out)
}

func parseCreatedTaskID(t *testing.T, output string) string {
	t.Helper()

	trimmed := strings.TrimSpace(output)
	prefix := "Created task "
	if !strings.HasPrefix(trimmed, prefix) {
		t.Fatalf("unexpected create output: %q", trimmed)
	}
	id := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	if id == "" {
		t.Fatalf("expected task ID in output: %q", trimmed)
	}
	return id
}

func createBoard(t *testing.T, repoRoot, storageRoot, name string) string {
	t.Helper()

	out, err := runMochiSticky(t, repoRoot, storageRoot, "board", "add", name)
	if err != nil {
		t.Fatalf("board add: %v", err)
	}
	return parseCreatedBoardID(t, out)
}

func parseCreatedBoardID(t *testing.T, output string) string {
	t.Helper()

	trimmed := strings.TrimSpace(output)
	prefix := "Created board "
	if !strings.HasPrefix(trimmed, prefix) {
		t.Fatalf("unexpected create board output: %q", trimmed)
	}
	id := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	if id == "" {
		t.Fatalf("expected board ID in output: %q", trimmed)
	}
	return id
}

func readTask(t *testing.T, storageRoot, id string) board.Task {
	t.Helper()

	taskPath := filepath.Join(storageRoot, "boards", "default", "tasks", id+".md")
	data, err := os.ReadFile(taskPath)
	if err != nil {
		t.Fatalf("read task file: %v", err)
	}
	task, err := (&board.Parser{}).Parse(data)
	if err != nil {
		t.Fatalf("parse task file: %v", err)
	}
	return task
}

func stripANSI(value string) string {
	return ansiRegexp.ReplaceAllString(value, "")
}

func setBoardContext(t *testing.T, repoRoot, storageRoot, boardID string, ctx board.BoardContext) {
	t.Helper()

	repo, err := board.NewRepositoryForBoardWithStorage(repoRoot, boardID, storageRoot)
	if err != nil {
		t.Fatalf("load board repo: %v", err)
	}
	if err := repo.UpdateBoardContextContext(context.Background(), ctx); err != nil {
		t.Fatalf("update board context: %v", err)
	}
}
