package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureInDir verifies that path stays within baseDir.
func EnsureInDir(baseDir, path string) error {
	rel, err := filepath.Rel(baseDir, path)
	if err != nil {
		return fmt.Errorf("shared: failed to resolve path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("shared: %w", ErrInvalidPath)
	}
	return nil
}

// PathExists reports whether a path exists on disk.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("shared: failed to stat path %s: %w", path, err)
}

// IsSubpath reports whether target is within base (or equal to base).
func IsSubpath(base, target string) bool {
	if strings.TrimSpace(base) == "" || strings.TrimSpace(target) == "" {
		return false
	}
	baseClean := filepath.Clean(base)
	targetClean := filepath.Clean(target)
	rel, err := filepath.Rel(baseClean, targetClean)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}
