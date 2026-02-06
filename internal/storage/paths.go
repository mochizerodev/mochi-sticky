package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigPaths captures overrides for config file locations.
type ConfigPaths struct {
	Boards    string `yaml:"boards"`
	ADR       string `yaml:"adr"`
	WikiIndex string `yaml:"wiki_index"`
}

// ResolvedConfigPaths contains resolved config file locations.
type ResolvedConfigPaths struct {
	Boards    string
	ADR       string
	WikiIndex string
}

// ResolveConfigPaths determines config file locations from config and defaults.
func ResolveConfigPaths(workingDir, storageRoot string, cfg Config) (ResolvedConfigPaths, error) {
	if strings.TrimSpace(workingDir) == "" {
		return ResolvedConfigPaths{}, fmt.Errorf("storage: working directory is required")
	}
	if strings.TrimSpace(storageRoot) == "" {
		return ResolvedConfigPaths{}, fmt.Errorf("storage: storage root is required")
	}

	boardsPath, err := resolveConfigPath(storageRoot, fallback(cfg.Paths.Boards, filepath.Join(storageRoot, "boards", "boards.yaml")), true)
	if err != nil {
		return ResolvedConfigPaths{}, err
	}
	adrPath, err := resolveConfigPath(storageRoot, fallback(cfg.Paths.ADR, filepath.Join(storageRoot, "adrs", "config.yaml")), true)
	if err != nil {
		return ResolvedConfigPaths{}, err
	}
	wikiIndexPath, err := resolveConfigPath(storageRoot, fallback(cfg.Paths.WikiIndex, filepath.Join(storageRoot, "wiki", "_index.yaml")), true)
	if err != nil {
		return ResolvedConfigPaths{}, err
	}

	return ResolvedConfigPaths{
		Boards:    boardsPath,
		ADR:       adrPath,
		WikiIndex: wikiIndexPath,
	}, nil
}

func resolveConfigPath(storageRoot, value string, allowMissing bool) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("storage: config path is required")
	}
	if !filepath.IsAbs(trimmed) {
		trimmed = filepath.Join(storageRoot, trimmed)
	}
	absPath, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("storage: failed to resolve config path: %w", err)
	}
	if allowMissing {
		if _, err := os.Stat(absPath); err != nil {
			if os.IsNotExist(err) {
				return absPath, nil
			}
			return "", fmt.Errorf("storage: failed to stat config path %s: %w", absPath, err)
		}
	}
	return absPath, nil
}
