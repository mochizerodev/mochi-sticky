package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureInDir(t *testing.T) {
	// Arrange
	base := t.TempDir()
	child := filepath.Join(base, "child", "file.txt")
	other := filepath.Join(t.TempDir(), "other.txt")

	// Act
	childErr := EnsureInDir(base, child)
	otherErr := EnsureInDir(base, other)

	// Assert
	if childErr != nil {
		t.Fatalf("expected child path to be allowed: %v", childErr)
	}
	if otherErr == nil {
		t.Fatalf("expected path outside base to be rejected")
	}
}

func TestPathExists(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	missing := filepath.Join(dir, "missing.txt")

	// Act
	exists, err := PathExists(file)
	missingExists, missingErr := PathExists(missing)

	// Assert
	if err != nil {
		t.Fatalf("path exists error: %v", err)
	}
	if !exists {
		t.Fatalf("expected file to exist")
	}
	if missingErr != nil {
		t.Fatalf("path exists missing error: %v", missingErr)
	}
	if missingExists {
		t.Fatalf("expected missing file to return false")
	}
}
