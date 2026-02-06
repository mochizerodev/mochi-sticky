package board

import (
	"context"
	"testing"
)

func TestLoadConfigContextDefault(t *testing.T) {
	repo, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}
	cfg, err := repo.LoadConfigContext(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Columns) == 0 {
		t.Fatalf("expected default columns")
	}
}

func TestUpdateBoardContextContext(t *testing.T) {
	repo, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}
	ctx := BoardContext{Scope: "Sprint", Owners: []string{"Alice"}}
	if err := repo.UpdateBoardContextContext(context.Background(), ctx); err != nil {
		t.Fatalf("update context: %v", err)
	}
	cfg, err := repo.LoadConfigContext(context.Background())
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Context.Scope != "Sprint" || len(cfg.Context.Owners) != 1 {
		t.Fatalf("expected updated context, got %+v", cfg.Context)
	}
}
