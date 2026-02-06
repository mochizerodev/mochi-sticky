package adr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_DefaultWhenMissing(t *testing.T) {
	root := t.TempDir()
	cfg, err := loadConfig(root)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.NextID != 1 {
		t.Fatalf("expected NextID 1, got %d", cfg.NextID)
	}
	if cfg.ConfigVersion != 1 {
		t.Fatalf("expected ConfigVersion 1, got %d", cfg.ConfigVersion)
	}
	if len(cfg.Columns) == 0 {
		t.Fatalf("expected default columns")
	}
}

func TestSaveConfig_RoundTrip(t *testing.T) {
	root := t.TempDir()
	cfg := Config{
		ConfigVersion: 1,
		NextID:        7,
		Columns: []Column{
			{Key: "proposed", Title: "Proposed"},
			{Key: "accepted", Title: "Accepted"},
		},
	}
	if err := saveConfig(root, cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	path := filepath.Join(root, ConfigFileName)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file: %v", err)
	}
	loaded, err := loadConfig(root)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if loaded.NextID != 7 {
		t.Fatalf("expected NextID 7, got %d", loaded.NextID)
	}
	if loaded.ConfigVersion != 1 {
		t.Fatalf("expected ConfigVersion 1, got %d", loaded.ConfigVersion)
	}
	if len(loaded.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(loaded.Columns))
	}
}
