package board

import (
	"context"
	"strings"
	"testing"
)

func TestBoardDescriptionRoundTrip(t *testing.T) {
	repo, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	if err := repo.UpdateBoardDescriptionContext(context.Background(), "Hello\n"); err != nil {
		t.Fatalf("update description: %v", err)
	}
	desc, err := repo.LoadBoardDescriptionContext(context.Background())
	if err != nil {
		t.Fatalf("load description: %v", err)
	}
	if strings.TrimSpace(desc) != "Hello" {
		t.Fatalf("expected description, got %q", desc)
	}

	if err := repo.UpdateBoardDescriptionContext(context.Background(), ""); err != nil {
		t.Fatalf("clear description: %v", err)
	}
	desc, err = repo.LoadBoardDescriptionContext(context.Background())
	if err != nil {
		t.Fatalf("load description: %v", err)
	}
	if strings.TrimSpace(desc) != "" {
		t.Fatalf("expected empty description, got %q", desc)
	}
}
