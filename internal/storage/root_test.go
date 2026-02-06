package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootPrecedence(t *testing.T) {
	// Arrange
	baseDir := t.TempDir()
	overrideDir := filepath.Join(baseDir, "override")
	envDir := filepath.Join(baseDir, "env")
	configDir := filepath.Join(baseDir, "config")
	defaultDir := filepath.Join(baseDir, DefaultStoreDir)

	for _, dir := range []string{overrideDir, envDir, configDir, defaultDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	configPath := filepath.Join(baseDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("storage_root: config\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv(EnvVar, "env")

	// Act
	overrideResolved, overrideErr := ResolveRoot(baseDir, false, "override")
	envResolved, envErr := ResolveRoot(baseDir, false, "")
	if err := os.Unsetenv(EnvVar); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	configResolved, configErr := ResolveRoot(baseDir, false, "")
	if err := os.Remove(configPath); err != nil {
		t.Fatalf("remove config: %v", err)
	}
	defaultResolved, defaultErr := ResolveRoot(baseDir, false, "")

	// Assert
	if overrideErr != nil {
		t.Fatalf("resolve override: %v", overrideErr)
	}
	if overrideResolved != overrideDir {
		t.Fatalf("expected override dir %q, got %q", overrideDir, overrideResolved)
	}
	if envErr != nil {
		t.Fatalf("resolve env: %v", envErr)
	}
	if envResolved != envDir {
		t.Fatalf("expected env dir %q, got %q", envDir, envResolved)
	}
	if configErr != nil {
		t.Fatalf("resolve config: %v", configErr)
	}
	if configResolved != configDir {
		t.Fatalf("expected config dir %q, got %q", configDir, configResolved)
	}
	if defaultErr != nil {
		t.Fatalf("resolve default: %v", defaultErr)
	}
	if defaultResolved != defaultDir {
		t.Fatalf("expected default dir %q, got %q", defaultDir, defaultResolved)
	}
}

func TestResolveRootAllowMissing(t *testing.T) {
	// Arrange
	baseDir := t.TempDir()
	missing := filepath.Join(baseDir, "missing-dir")

	// Act
	resolved, allowErr := ResolveRoot(baseDir, true, "missing-dir")
	_, disallowErr := ResolveRoot(baseDir, false, "missing-dir")

	// Assert
	if allowErr != nil {
		t.Fatalf("expected allowMissing to succeed, got %v", allowErr)
	}
	if resolved != missing {
		t.Fatalf("expected %q, got %q", missing, resolved)
	}
	if disallowErr == nil {
		t.Fatalf("expected error when allowMissing=false")
	}
}

func TestResolveRootRejectsFile(t *testing.T) {
	// Arrange
	baseDir := t.TempDir()
	filePath := filepath.Join(baseDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Act
	_, err := ResolveRoot(baseDir, false, "file.txt")

	// Assert
	if err == nil {
		t.Fatalf("expected error for file path")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Arrange
	baseDir := t.TempDir()
	configDir := filepath.Join(baseDir, DefaultStoreDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", configDir, err)
	}
	configPath := filepath.Join(configDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("storage_root: ["), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Act
	_, err := LoadConfig(baseDir)

	// Assert
	if err == nil {
		t.Fatalf("expected invalid yaml error")
	}
}
