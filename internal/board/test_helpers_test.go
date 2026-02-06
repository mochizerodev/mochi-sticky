package board

import (
	"path/filepath"
	"testing"
)

func setupRepo(t *testing.T) (*Repository, string, string) {
	t.Helper()
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	repo, err := NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStore(); err != nil {
		t.Fatalf("init store: %v", err)
	}
	return repo, baseDir, storageRoot
}

func setupBoardRepo(t *testing.T) (*BoardRepository, string, string) {
	t.Helper()
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	repo, err := NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStore(); err != nil {
		t.Fatalf("init store: %v", err)
	}
	boardRepo, err := NewBoardRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new board repo: %v", err)
	}
	return boardRepo, baseDir, storageRoot
}
