package templates

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mochi-sticky/internal/storage"
)

func TestCopyDirIfEmptyCopiesContents(t *testing.T) {
	base := t.TempDir()
	source := filepath.Join(base, "source")
	target := filepath.Join(base, "target")

	if err := os.MkdirAll(filepath.Join(source, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "nested", "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	if err := copyDirIfEmpty(source, target); err != nil {
		t.Fatalf("copyDirIfEmpty: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(target, "nested", "file.txt"))
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected copied content, got %q", string(data))
	}
}

func TestCopyDirIfEmptySkipsWhenTargetHasEntries(t *testing.T) {
	base := t.TempDir()
	source := filepath.Join(base, "source")
	target := filepath.Join(base, "target")

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "file.txt"), []byte("source"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "existing.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write target file: %v", err)
	}

	if err := copyDirIfEmpty(source, target); err != nil {
		t.Fatalf("copyDirIfEmpty: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "file.txt")); err == nil {
		t.Fatalf("expected source file to be skipped when target not empty")
	}
}

func TestCopyEmbeddedDirIfEmptyCopiesAssets(t *testing.T) {
	target := t.TempDir()

	if err := copyEmbeddedDirIfEmpty("assets/task", target); err != nil {
		t.Fatalf("copyEmbeddedDirIfEmpty: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "default.md")); err != nil {
		t.Fatalf("expected embedded asset to be copied: %v", err)
	}
}

func TestCopyEmbeddedFileIfMissingCopiesTemplate(t *testing.T) {
	target := filepath.Join(t.TempDir(), "default.md")

	if err := copyEmbeddedFileIfMissing("assets/task/default.md", target); err != nil {
		t.Fatalf("copyEmbeddedFileIfMissing: %v", err)
	}

	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected embedded file to be copied: %v", err)
	}
}

func TestSeedDefaultsContextCopiesTemplates(t *testing.T) {
	base := t.TempDir()
	storageRoot := filepath.Join(base, "store")
	paths := storage.TemplatePaths{
		Root:    filepath.Join(storageRoot, "templates"),
		ADR:     filepath.Join(storageRoot, "templates", "adr"),
		Task:    filepath.Join(storageRoot, "templates", "task"),
		Board:   filepath.Join(storageRoot, "templates", "board"),
		Wiki:    filepath.Join(storageRoot, "templates", "wiki"),
		WikiPDF: filepath.Join(storageRoot, "templates", "wiki", "wiki_pdf_template.tex"),
	}

	if err := SeedDefaultsContext(context.Background(), storageRoot, paths); err != nil {
		t.Fatalf("SeedDefaultsContext: %v", err)
	}

	checks := []string{
		filepath.Join(paths.ADR, "default.md"),
		filepath.Join(paths.Task, "default.md"),
		filepath.Join(paths.Board, "kanban.yaml"),
		filepath.Join(paths.Wiki, "architecture-overview.md"),
	}
	for _, path := range checks {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected seeded file %s: %v", path, err)
		}
	}
	if _, err := os.Stat(paths.WikiPDF); err == nil {
		return
	}
}

func TestSeedDefaultsContextHonorsCancel(t *testing.T) {
	base := t.TempDir()
	paths := storage.TemplatePaths{
		ADR:   filepath.Join(base, "adr"),
		Task:  filepath.Join(base, "task"),
		Board: filepath.Join(base, "board"),
		Wiki:  filepath.Join(base, "wiki"),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := SeedDefaultsContext(ctx, base, paths); err == nil {
		t.Fatalf("expected cancellation error")
	}
}

func TestCopyEmbeddedDirIfEmptyInvalidSource(t *testing.T) {
	if err := copyEmbeddedDirIfEmpty("wrong/prefix", t.TempDir()); err == nil {
		t.Fatalf("expected error for invalid embedded source")
	}
}

func TestCopyEmbeddedFileIfMissingInvalidSource(t *testing.T) {
	if err := copyEmbeddedFileIfMissing("wrong/prefix", filepath.Join(t.TempDir(), "file")); err == nil {
		t.Fatalf("expected error for invalid embedded source")
	}
}

func TestCopyEmbeddedDirIfEmptyMissingSource(t *testing.T) {
	target := t.TempDir()
	if err := copyEmbeddedDirIfEmpty("assets/does-not-exist", target); err != nil {
		t.Fatalf("expected missing embedded source to be ignored: %v", err)
	}
}

func TestSeedDefaultsContextRejectsEmptyRoot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := SeedDefaultsContext(ctx, "", storage.TemplatePaths{}); err == nil {
		t.Fatalf("expected error for empty storage root")
	}
}

func TestEnsureDirRejectsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := ensureDir(path); err == nil {
		t.Fatalf("expected error for file path")
	}
}
