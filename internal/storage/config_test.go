package storage

import "testing"

func TestSaveConfigToRootRoundTrip(t *testing.T) {
	root := t.TempDir()
	cfg := Config{
		Editor: "vim",
		Paths: ConfigPaths{
			Boards:    "boards/boards.yaml",
			ADR:       "adrs/config.yaml",
			WikiIndex: "wiki/_index.yaml",
		},
	}
	if err := SaveConfigToRoot(root, cfg); err != nil {
		t.Fatalf("SaveConfigToRoot: %v", err)
	}
	loaded, err := LoadConfigFromRoot(root)
	if err != nil {
		t.Fatalf("LoadConfigFromRoot: %v", err)
	}
	if loaded.Editor != "vim" {
		t.Fatalf("expected editor %q, got %q", "vim", loaded.Editor)
	}
	if loaded.Paths.Boards != "boards/boards.yaml" {
		t.Fatalf("expected boards path %q, got %q", "boards/boards.yaml", loaded.Paths.Boards)
	}
}
