package wiki

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIndexMissing(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	path := filepath.Join(dir, "_index.yaml")

	// Act
	_, err := LoadIndex(path)

	// Assert
	if !errors.Is(err, ErrIndexNotFound) {
		t.Fatalf("expected ErrIndexNotFound, got %v", err)
	}
}

func TestSaveAndLoadIndex(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	path := filepath.Join(dir, "_index.yaml")

	index := Index{
		Sections: []IndexSection{
			{Title: "Architecture", Slug: "architecture", Order: 1, Pages: []string{"overview", "decisions"}},
		},
	}

	// Act
	saveErr := SaveIndex(path, index)
	loaded, loadErr := LoadIndex(path)
	_, statErr := os.Stat(path)

	// Assert
	if saveErr != nil {
		t.Fatalf("failed to save index: %v", saveErr)
	}
	if loadErr != nil {
		t.Fatalf("failed to load index: %v", loadErr)
	}
	if loaded.IndexVersion != 1 {
		t.Fatalf("expected IndexVersion 1, got %d", loaded.IndexVersion)
	}
	if len(loaded.Sections) != 1 || loaded.Sections[0].Slug != "architecture" {
		t.Fatalf("unexpected index: %+v", loaded)
	}
	if statErr != nil {
		t.Fatalf("expected index file to exist: %v", statErr)
	}
}
