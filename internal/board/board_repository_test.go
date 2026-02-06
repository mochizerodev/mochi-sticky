package board

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveBoardUpdatesActive(t *testing.T) {
	// Arrange
	boardRepo, _, _ := setupBoardRepo(t)

	created, err := boardRepo.CreateBoard("Work")
	if err != nil {
		t.Fatalf("create board: %v", err)
	}
	if err := boardRepo.SetActiveBoard(created.ID); err != nil {
		t.Fatalf("set active board: %v", err)
	}

	// Act
	archived, archiveErr := boardRepo.ArchiveBoard(created.ID)
	registry, registryErr := boardRepo.LoadRegistry()

	// Assert
	if archiveErr != nil {
		t.Fatalf("archive board: %v", archiveErr)
	}
	if registryErr != nil {
		t.Fatalf("load registry: %v", registryErr)
	}
	if !archived.Archived {
		t.Fatalf("expected archived flag to be true")
	}
	if registry.Active == created.ID {
		t.Fatalf("expected active board to change after archiving active")
	}
	if registry.Active == "" {
		t.Fatalf("expected active board to be set")
	}
}

func TestDeleteBoardConstraints(t *testing.T) {
	// Arrange
	boardRepo, _, _ := setupBoardRepo(t)

	// Act
	err := boardRepo.DeleteBoard("default")

	// Assert
	if err == nil {
		t.Fatalf("expected delete to be forbidden when only one board exists")
	}
	if !errors.Is(err, ErrBoardDeleteForbidden) {
		t.Fatalf("expected ErrBoardDeleteForbidden, got %v", err)
	}
}

func TestDeleteBoardRemovesDirectory(t *testing.T) {
	// Arrange
	boardRepo, _, storageRoot := setupBoardRepo(t)

	created, err := boardRepo.CreateBoard("Work")
	if err != nil {
		t.Fatalf("create board: %v", err)
	}
	if _, err := boardRepo.ArchiveBoard(created.ID); err != nil {
		t.Fatalf("archive board: %v", err)
	}

	// Act
	deleteErr := boardRepo.DeleteBoard(created.ID)
	boardDir := filepath.Join(storageRoot, "boards", created.ID)

	// Assert
	if deleteErr != nil {
		t.Fatalf("delete board: %v", deleteErr)
	}
	if _, err := os.Stat(boardDir); err == nil {
		t.Fatalf("expected board directory to be removed")
	}
}

func TestResolveBoardPathsInvalid(t *testing.T) {
	// Arrange
	boardRepo, _, _ := setupBoardRepo(t)

	// Act
	_, _, _, _, err := boardRepo.ResolveBoardPaths("missing")

	// Assert
	if err == nil {
		t.Fatalf("expected error for missing board")
	}
	if !errors.Is(err, ErrBoardNotFound) {
		t.Fatalf("expected ErrBoardNotFound, got %v", err)
	}
}
