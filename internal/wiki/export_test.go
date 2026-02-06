package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportMarkdown(t *testing.T) {
	// Arrange
	root := filepath.Join(t.TempDir(), "wiki")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("failed to create root: %v", err)
	}
	page1 := `---
title: "One"
slug: "one"
---
First page
`
	page2 := `---
title: "Two"
slug: "two"
---
Second page
`
	if err := os.WriteFile(filepath.Join(root, "one.md"), []byte(page1), 0o644); err != nil {
		t.Fatalf("failed to write page1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "two.md"), []byte(page2), 0o644); err != nil {
		t.Fatalf("failed to write page2: %v", err)
	}

	manifest := []ManifestEntry{
		{Title: "One", Slug: "one"},
		{Title: "Two", Slug: "two"},
	}

	// Act
	exported, err := ExportMarkdown(root, manifest)

	// Assert
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	out := string(exported)
	if !strings.Contains(out, "# One") || !strings.Contains(out, "# Two") {
		t.Fatalf("missing headings in export: %s", out)
	}
	if !strings.Contains(out, "---") {
		t.Fatalf("missing page break: %s", out)
	}
}

func TestExportMarkdownMulti(t *testing.T) {
	// Arrange
	rootA := filepath.Join(t.TempDir(), "wiki-a")
	rootB := filepath.Join(t.TempDir(), "wiki-b")
	if err := os.MkdirAll(rootA, 0o755); err != nil {
		t.Fatalf("failed to create rootA: %v", err)
	}
	if err := os.MkdirAll(rootB, 0o755); err != nil {
		t.Fatalf("failed to create rootB: %v", err)
	}

	if err := os.WriteFile(filepath.Join(rootA, "a.md"), []byte(`---
title: "A"
slug: "a"
---
Alpha
`), 0o644); err != nil {
		t.Fatalf("failed to write A: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootB, "b.md"), []byte(`---
title: "B"
slug: "b"
---
Beta
`), 0o644); err != nil {
		t.Fatalf("failed to write B: %v", err)
	}

	manifests := []RootManifest{
		{Root: rootA, Prefix: "core", Pages: []ManifestEntry{{Title: "A", Slug: "a"}}},
		{Root: rootB, Prefix: "ext", Pages: []ManifestEntry{{Title: "B", Slug: "b"}}},
	}

	// Act
	out, err := ExportMarkdownMulti(manifests)

	// Assert
	if err != nil {
		t.Fatalf("export multi failed: %v", err)
	}
	if !strings.Contains(string(out), "# A") || !strings.Contains(string(out), "# B") {
		t.Fatalf("missing headings in multi export: %s", string(out))
	}
}
