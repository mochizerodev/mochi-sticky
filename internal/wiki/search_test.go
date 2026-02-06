package wiki

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchPages(t *testing.T) {
	// Arrange
	root := filepath.Join(t.TempDir(), "wiki")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("failed to create root: %v", err)
	}
	pagePath := filepath.Join(root, "home.md")
	content := `---
title: "Home"
slug: "home"
---
Hello wiki
Another line
`
	if err := os.WriteFile(pagePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write page: %v", err)
	}

	// Act
	results, err := SearchPages(root, SearchOptions{Query: "Hello", CaseInsensitive: true})

	// Assert
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "home" || results[0].Line != 5 {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}
