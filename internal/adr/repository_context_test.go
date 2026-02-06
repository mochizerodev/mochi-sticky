package adr

import (
	"context"
	"testing"
	"time"
)

func TestCreateADRContextAndStatusUpdate(t *testing.T) {
	root := t.TempDir()
	repo, err := NewRepository(root)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	record, err := repo.CreateADRContext(context.Background(), "First ADR", CreateOptions{
		Status: "proposed",
		Date:   time.Now(),
	})
	if err != nil {
		t.Fatalf("create adr: %v", err)
	}

	if err := repo.UpdateADRStatusContext(context.Background(), record.ID, "accepted"); err != nil {
		t.Fatalf("update status: %v", err)
	}
	adrs, err := repo.ListADRsContext(context.Background())
	if err != nil {
		t.Fatalf("list adrs: %v", err)
	}
	if len(adrs) != 1 || adrs[0].Status != "accepted" {
		t.Fatalf("expected updated status, got %+v", adrs)
	}
}

func TestListADRsContextCanceled(t *testing.T) {
	root := t.TempDir()
	repo, err := NewRepository(root)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListADRsContext(ctx); err == nil {
		t.Fatalf("expected canceled error")
	}
}
