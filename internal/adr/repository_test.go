package adr

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRepository_CreateListGetUpdateStatus(t *testing.T) {
	root := filepath.Join(t.TempDir(), "adrs")
	repo, err := NewRepository(root)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	if err := repo.InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	created, err := repo.CreateADR("Decision One", CreateOptions{
		Status: "proposed",
		Date:   time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC),
		Tags:   []string{"test"},
	})
	if err != nil {
		t.Fatalf("CreateADR: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected id 1, got %d", created.ID)
	}
	if _, err := os.Stat(created.FilePath); err != nil {
		t.Fatalf("expected file on disk: %v", err)
	}

	list, err := repo.ListADRs()
	if err != nil {
		t.Fatalf("ListADRs: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 adr, got %d", len(list))
	}

	got, err := repo.GetADRByID(1)
	if err != nil {
		t.Fatalf("GetADRByID: %v", err)
	}
	if got.Title != "Decision One" {
		t.Fatalf("expected title, got %q", got.Title)
	}

	if err := repo.UpdateADRStatus(1, "accepted"); err != nil {
		t.Fatalf("UpdateADRStatus: %v", err)
	}
	updated, err := repo.GetADRByID(1)
	if err != nil {
		t.Fatalf("GetADRByID: %v", err)
	}
	if updated.Status != "accepted" {
		t.Fatalf("expected status accepted, got %q", updated.Status)
	}
}

func TestRepository_DeleteADR(t *testing.T) {
	root := filepath.Join(t.TempDir(), "adrs")
	repo, err := NewRepository(root)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	if err := repo.InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	created, err := repo.CreateADR("Decision Delete", CreateOptions{
		Status: "proposed",
		Date:   time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("CreateADR: %v", err)
	}

	if err := repo.DeleteADR(created.ID); err != nil {
		t.Fatalf("DeleteADR: %v", err)
	}
	if _, err := os.Stat(created.FilePath); !os.IsNotExist(err) {
		t.Fatalf("expected adr file deleted, got err=%v", err)
	}
	if _, err := repo.GetADRByID(created.ID); !errors.Is(err, ErrADRNotFound) {
		t.Fatalf("expected ErrADRNotFound, got %v", err)
	}
}
