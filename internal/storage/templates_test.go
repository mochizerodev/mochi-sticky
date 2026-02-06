package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTemplatesDefaults(t *testing.T) {
	base := t.TempDir()
	storageRoot := filepath.Join(base, "store")
	if err := os.MkdirAll(storageRoot, 0o755); err != nil {
		t.Fatalf("mkdir storageRoot: %v", err)
	}
	paths, err := ResolveTemplates(base, storageRoot, Config{})
	if err != nil {
		t.Fatalf("ResolveTemplates: %v", err)
	}
	if paths.Root == "" || paths.ADR == "" || paths.Task == "" || paths.Board == "" || paths.Wiki == "" || paths.WikiPDF == "" {
		t.Fatalf("expected non-empty paths: %+v", paths)
	}
	if filepath.Dir(paths.ADR) != filepath.Join(storageRoot, "templates") {
		t.Fatalf("expected ADR path under templates root, got %s", paths.ADR)
	}
}

func TestResolveTemplatesOverrides(t *testing.T) {
	base := t.TempDir()
	storageRoot := filepath.Join(base, "store")
	if err := os.MkdirAll(storageRoot, 0o755); err != nil {
		t.Fatalf("mkdir storageRoot: %v", err)
	}
	customRoot := filepath.Join(base, "custom-templates")
	customPDF := filepath.Join(customRoot, "pdf.tex")

	cfg := Config{
		PDFTemplate: customPDF,
		Templates: TemplatesConfig{
			Root:  customRoot,
			ADR:   filepath.Join(customRoot, "adr"),
			Task:  filepath.Join(customRoot, "task"),
			Board: filepath.Join(customRoot, "board"),
			Wiki:  filepath.Join(customRoot, "wiki"),
		},
	}
	paths, err := ResolveTemplates(base, storageRoot, cfg)
	if err != nil {
		t.Fatalf("ResolveTemplates: %v", err)
	}
	if paths.Root != customRoot {
		t.Fatalf("expected root %s, got %s", customRoot, paths.Root)
	}
	if paths.WikiPDF != customPDF {
		t.Fatalf("expected pdf %s, got %s", customPDF, paths.WikiPDF)
	}
}

func TestResolveTemplatesRejectsMissingInputs(t *testing.T) {
	if _, err := ResolveTemplates("", "root", Config{}); err == nil {
		t.Fatalf("expected error for empty workingDir")
	}
	if _, err := ResolveTemplates("work", "", Config{}); err == nil {
		t.Fatalf("expected error for empty storageRoot")
	}
}

func TestResolveTemplatePathTypeErrors(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := resolveTemplateDir(base, file, false); err == nil {
		t.Fatalf("expected error for dir expected but file provided")
	}
	dir := filepath.Join(base, "dir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}
	if _, err := resolveTemplateFile(base, dir, false); err == nil {
		t.Fatalf("expected error for file expected but dir provided")
	}
}

func TestResolveTemplatePathAllowMissing(t *testing.T) {
	base := t.TempDir()
	missing := filepath.Join(base, "missing")
	resolved, err := resolveTemplateDir(base, missing, true)
	if err != nil {
		t.Fatalf("expected allow missing, got %v", err)
	}
	if resolved == "" {
		t.Fatalf("expected resolved path")
	}
}
