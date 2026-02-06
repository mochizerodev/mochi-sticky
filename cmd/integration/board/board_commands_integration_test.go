package integration

import (
	"fmt"
	"strings"
	"testing"

	"mochi-sticky/internal/board"
)

func TestBoardAddCommandCreatesBoard(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)

	// Act
	boardID := createBoard(t, repoRoot, storageRoot, "Roadmap")

	// Assert
	if boardID == "" {
		t.Fatalf("expected board ID to be set")
	}
}

func TestBoardListCommandShowsBoards(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	boardID := createBoard(t, repoRoot, storageRoot, "Roadmap")

	// Act
	listOut, err := runMochiSticky(t, repoRoot, storageRoot, "board", "list")
	if err != nil {
		t.Fatalf("board list: %v", err)
	}

	// Assert
	if !strings.Contains(listOut, fmt.Sprintf("%s - Roadmap", boardID)) {
		t.Fatalf("expected board in list, got:\n%s", listOut)
	}
}

func TestBoardShowCommandPrintsContext(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	boardID := createBoard(t, repoRoot, storageRoot, "Roadmap")
	setBoardContext(t, repoRoot, storageRoot, boardID, board.BoardContext{
		Scope:   "Delivery",
		Release: "2026.02",
		Target:  "Q1",
		Owners:  []string{"Ada", "Linus"},
		Notes:   "Focus on reliability.",
	})

	// Act
	showOut, err := runMochiSticky(t, repoRoot, storageRoot, "board", "show", boardID)
	if err != nil {
		t.Fatalf("board show: %v", err)
	}

	// Assert
	if !strings.Contains(showOut, fmt.Sprintf("ID: %s", boardID)) {
		t.Fatalf("expected board ID in show output, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Name: Roadmap") {
		t.Fatalf("expected board name in show output, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Context:") {
		t.Fatalf("expected context block in show output, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Scope: Delivery") {
		t.Fatalf("expected scope in show output, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Release: 2026.02") {
		t.Fatalf("expected release in show output, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Target: Q1") {
		t.Fatalf("expected target in show output, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Owners: Ada, Linus") {
		t.Fatalf("expected owners in show output, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Notes: Focus on reliability.") {
		t.Fatalf("expected notes in show output, got:\n%s", showOut)
	}
}

func TestBoardRenameCommandUpdatesName(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	boardID := createBoard(t, repoRoot, storageRoot, "Roadmap")

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "board", "rename", boardID, "Product Roadmap"); err != nil {
		t.Fatalf("board rename: %v", err)
	}
	showOut, err := runMochiSticky(t, repoRoot, storageRoot, "board", "show", boardID)
	if err != nil {
		t.Fatalf("board show: %v", err)
	}

	// Assert
	if !strings.Contains(showOut, "Name: Product Roadmap") {
		t.Fatalf("expected renamed board in show output, got:\n%s", showOut)
	}
}

func TestBoardUseCommandSetsActiveBoard(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	boardID := createBoard(t, repoRoot, storageRoot, "Roadmap")

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "board", "use", boardID); err != nil {
		t.Fatalf("board use: %v", err)
	}
	listOut, err := runMochiSticky(t, repoRoot, storageRoot, "board", "list")
	if err != nil {
		t.Fatalf("board list: %v", err)
	}

	// Assert
	if !strings.Contains(listOut, fmt.Sprintf("* %s - Roadmap", boardID)) {
		t.Fatalf("expected active board in list, got:\n%s", listOut)
	}
}

func TestBoardArchiveCommandMarksBoard(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	boardID := createBoard(t, repoRoot, storageRoot, "Roadmap")

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "board", "archive", boardID, "--force"); err != nil {
		t.Fatalf("board archive: %v", err)
	}
	listOut, err := runMochiSticky(t, repoRoot, storageRoot, "board", "list")
	if err != nil {
		t.Fatalf("board list: %v", err)
	}

	// Assert
	if !strings.Contains(listOut, fmt.Sprintf("%s - Roadmap (archived)", boardID)) {
		t.Fatalf("expected archived board in list, got:\n%s", listOut)
	}
}

func TestBoardDeleteCommandRemovesBoard(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupStorage(t)
	boardID := createBoard(t, repoRoot, storageRoot, "Roadmap")
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "board", "archive", boardID, "--force"); err != nil {
		t.Fatalf("board archive: %v", err)
	}

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "board", "delete", boardID, "--force"); err != nil {
		t.Fatalf("board delete: %v", err)
	}
	listOut, err := runMochiSticky(t, repoRoot, storageRoot, "board", "list")
	if err != nil {
		t.Fatalf("board list: %v", err)
	}

	// Assert
	if strings.Contains(listOut, boardID) {
		t.Fatalf("expected deleted board to be removed, got:\n%s", listOut)
	}
}
