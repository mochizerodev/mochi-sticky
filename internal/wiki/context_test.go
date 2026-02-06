package wiki

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoadAndSaveIndexContext(t *testing.T) {
	root := t.TempDir()
	indexPath := filepath.Join(root, "_index.yaml")
	index := Index{Sections: []IndexSection{{Title: "General", Slug: "", Pages: []string{"alpha"}}}}
	if err := SaveIndexContext(context.Background(), indexPath, index); err != nil {
		t.Fatalf("save index: %v", err)
	}
	loaded, err := LoadIndexContext(context.Background(), indexPath)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}
	if len(loaded.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(loaded.Sections))
	}
}

func TestBuildNavTreeContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := BuildNavTreeContext(ctx, Index{}, nil); err == nil {
		t.Fatalf("expected canceled error")
	}
}

func TestListPagesContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ListPagesContext(ctx, t.TempDir()); err == nil {
		t.Fatalf("expected canceled error")
	}
}

func TestGenerateIndexContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GenerateIndexContext(ctx, nil); err == nil {
		t.Fatalf("expected canceled error")
	}
}

func TestBuildRootManifestContext(t *testing.T) {
	root := t.TempDir()
	page := Page{Title: "Alpha", Slug: "alpha", Content: "Body"}
	if err := SavePage(filepath.Join(root, "alpha.md"), page); err != nil {
		t.Fatalf("save page: %v", err)
	}
	manifest, err := BuildRootManifestContext(context.Background(), root, "")
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if len(manifest.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(manifest.Pages))
	}
}

func TestLoadIndexContextMissing(t *testing.T) {
	_, err := LoadIndexContext(context.Background(), filepath.Join(t.TempDir(), "_index.yaml"))
	if err == nil {
		t.Fatalf("expected ErrIndexNotFound")
	}
	if err != ErrIndexNotFound {
		t.Fatalf("expected ErrIndexNotFound, got %v", err)
	}
}

func TestSaveIndexContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := SaveIndexContext(ctx, filepath.Join(t.TempDir(), "_index.yaml"), Index{})
	if err == nil {
		t.Fatalf("expected canceled error")
	}
}

func TestListPagesContextWithMissingRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	pages, err := ListPagesContext(context.Background(), root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("expected no pages for missing root")
	}
}

func TestListPagesFromIndexContext(t *testing.T) {
	root := t.TempDir()
	if err := SavePage(filepath.Join(root, "alpha.md"), Page{
		Title:   "Alpha",
		Slug:    "alpha",
		Content: "Alpha body",
	}); err != nil {
		t.Fatalf("save page: %v", err)
	}
	index := Index{Sections: []IndexSection{{Title: "General", Slug: "", Pages: []string{"alpha"}}}}
	pages, err := ListPagesFromIndexContext(context.Background(), root, index)
	if err != nil {
		t.Fatalf("list pages: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
}
