package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	EnvVar          = "MOCHI_STICKY_STORAGE"
	ConfigFileName  = "mochi-sticky.yaml"
	DefaultStoreDir = ".sticky"
)

// Config captures storage root configuration.
type Config struct {
	StorageRoot string          `yaml:"storage_root"`
	PDFTemplate string          `yaml:"pdf_template"`
	Editor      string          `yaml:"editor"`
	Templates   TemplatesConfig `yaml:"templates"`
	Paths       ConfigPaths     `yaml:"config_paths"`
}

// ResolveRoot determines the storage root based on override, env, config, or defaults.
func ResolveRoot(workingDir string, allowMissing bool, override string) (string, error) {
	if strings.TrimSpace(workingDir) == "" {
		return "", fmt.Errorf("storage: working directory is required")
	}

	if strings.TrimSpace(override) != "" {
		return resolvePath(workingDir, override, allowMissing)
	}

	if env := strings.TrimSpace(os.Getenv(EnvVar)); env != "" {
		return resolvePath(workingDir, env, allowMissing)
	}

	legacyConfigPath := filepath.Join(workingDir, ConfigFileName)
	if _, err := os.Stat(legacyConfigPath); err == nil {
		config, err := loadConfigAtPath(legacyConfigPath)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(config.StorageRoot) != "" {
			return resolvePath(workingDir, config.StorageRoot, allowMissing)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("storage: failed to stat config %s: %w", legacyConfigPath, err)
	}

	return resolvePath(workingDir, DefaultStoreDir, allowMissing)
}

// ResolvePath resolves a storage path relative to workingDir.
func ResolvePath(workingDir, value string, allowMissing bool) (string, error) {
	return resolvePath(workingDir, value, allowMissing)
}

func resolvePath(workingDir, value string, allowMissing bool) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("storage: path is required")
	}

	if !filepath.IsAbs(trimmed) {
		trimmed = filepath.Join(workingDir, trimmed)
	}

	absPath, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("storage: failed to resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			if allowMissing {
				return absPath, nil
			}
			return "", fmt.Errorf("storage: path does not exist: %s", absPath)
		}
		return "", fmt.Errorf("storage: failed to stat path %s: %w", absPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("storage: path is not a directory: %s", absPath)
	}
	return absPath, nil
}

// LoadConfig reads the storage config file from disk.
func LoadConfig(workingDir string) (Config, error) {
	return LoadConfigWithOverride(workingDir, "")
}

// LoadConfigWithOverride reads the storage config from the override/env/default root.
func LoadConfigWithOverride(workingDir, override string) (Config, error) {
	if strings.TrimSpace(workingDir) == "" {
		return Config{}, fmt.Errorf("storage: working directory is required")
	}

	if strings.TrimSpace(override) != "" {
		root, err := resolvePath(workingDir, override, true)
		if err != nil {
			return Config{}, err
		}
		return LoadConfigFromRoot(root)
	}

	if env := strings.TrimSpace(os.Getenv(EnvVar)); env != "" {
		root, err := resolvePath(workingDir, env, true)
		if err != nil {
			return Config{}, err
		}
		return LoadConfigFromRoot(root)
	}

	defaultRoot := filepath.Join(workingDir, DefaultStoreDir)
	cfg, err := LoadConfigFromRoot(defaultRoot)
	if err != nil {
		return Config{}, err
	}
	if !isEmptyConfig(cfg) {
		return cfg, nil
	}

	legacyPath := filepath.Join(workingDir, ConfigFileName)
	if _, err := os.Stat(legacyPath); err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("storage: failed to stat config %s: %w", legacyPath, err)
	}
	return loadConfigAtPath(legacyPath)
}

// LoadConfigFromRoot reads the storage config from a specific storage root.
func LoadConfigFromRoot(storageRoot string) (Config, error) {
	if strings.TrimSpace(storageRoot) == "" {
		return Config{}, fmt.Errorf("storage: storage root is required")
	}
	configPath := filepath.Join(storageRoot, ConfigFileName)
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("storage: failed to stat config %s: %w", configPath, err)
	}
	return loadConfigAtPath(configPath)
}

// SaveConfigToRoot writes the storage config to a specific storage root.
func SaveConfigToRoot(storageRoot string, cfg Config) error {
	if strings.TrimSpace(storageRoot) == "" {
		return fmt.Errorf("storage: storage root is required")
	}
	if err := os.MkdirAll(storageRoot, 0o755); err != nil {
		return fmt.Errorf("storage: failed to create storage root %s: %w", storageRoot, err)
	}
	configPath := filepath.Join(storageRoot, ConfigFileName)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("storage: failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("storage: failed to write config %s: %w", configPath, err)
	}
	return nil
}

func loadConfigAtPath(configPath string) (Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("storage: failed to read config %s: %w", configPath, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("storage: failed to parse config %s: %w", configPath, err)
	}
	return cfg, nil
}

func isEmptyConfig(cfg Config) bool {
	return strings.TrimSpace(cfg.StorageRoot) == "" &&
		strings.TrimSpace(cfg.PDFTemplate) == "" &&
		strings.TrimSpace(cfg.Editor) == "" &&
		isEmptyTemplatesConfig(cfg.Templates) &&
		isEmptyConfigPaths(cfg.Paths)
}

func isEmptyTemplatesConfig(cfg TemplatesConfig) bool {
	return strings.TrimSpace(cfg.Root) == "" &&
		strings.TrimSpace(cfg.ADR) == "" &&
		strings.TrimSpace(cfg.Task) == "" &&
		strings.TrimSpace(cfg.Board) == "" &&
		strings.TrimSpace(cfg.Wiki) == "" &&
		strings.TrimSpace(cfg.WikiPDF) == ""
}

func isEmptyConfigPaths(cfg ConfigPaths) bool {
	return strings.TrimSpace(cfg.Boards) == "" &&
		strings.TrimSpace(cfg.ADR) == "" &&
		strings.TrimSpace(cfg.WikiIndex) == ""
}
