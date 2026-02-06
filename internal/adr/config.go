package adr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/shared"

	"gopkg.in/yaml.v3"
)

const (
	// ConfigFileName is the ADR status/config filename within the ADR root.
	ConfigFileName = "config.yaml"
	// TemplatesDirName is the directory name for ADR templates within the ADR root.
	TemplatesDirName = "templates"
)

// Column defines a single ADR status column.
type Column struct {
	Key   string `yaml:"key"`
	Title string `yaml:"title"`
}

// Config defines ADR storage configuration.
type Config struct {
	ConfigVersion int      `yaml:"config_version"`
	NextID        int      `yaml:"next_id"`
	Columns       []Column `yaml:"columns"`
}

// DefaultConfig returns the default ADR configuration.
func DefaultConfig() Config {
	return Config{
		ConfigVersion: 1,
		NextID:        1,
		Columns: []Column{
			{Key: "proposed", Title: "Proposed"},
			{Key: "accepted", Title: "Accepted"},
			{Key: "rejected", Title: "Rejected"},
			{Key: "deprecated", Title: "Deprecated"},
			{Key: "superseded", Title: "Superseded"},
		},
	}
}

// RenderConfig marshals the default config into YAML bytes.
func RenderConfig() ([]byte, error) {
	data, err := yaml.Marshal(DefaultConfig())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func configPath(root string) string {
	return filepath.Join(root, ConfigFileName)
}

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.ConfigVersion <= 0 {
		cfg.ConfigVersion = 1
	}
	if cfg.NextID <= 0 {
		cfg.NextID = 1
	}
	if len(cfg.Columns) == 0 {
		cfg.Columns = DefaultConfig().Columns
	}

	clean := make([]Column, 0, len(cfg.Columns))
	seen := make(map[string]struct{})
	for _, column := range cfg.Columns {
		key := strings.TrimSpace(column.Key)
		if key == "" {
			continue
		}
		lower := strings.ToLower(key)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		title := strings.TrimSpace(column.Title)
		clean = append(clean, Column{
			Key:   key,
			Title: title,
		})
	}
	if len(clean) == 0 {
		return Config{}, fmt.Errorf("adr: %w", ErrInvalidConfig)
	}
	cfg.Columns = clean
	return cfg, nil
}

func loadConfig(root string) (Config, error) {
	return loadConfigContext(context.Background(), root)
}

func loadConfigContext(ctx context.Context, root string) (Config, error) {
	path := configPath(root)
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err()
	default:
	}
	if err := shared.EnsureInDir(root, path); err != nil {
		return Config{}, err
	}
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err()
	default:
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return normalizeConfig(DefaultConfig())
		}
		return Config{}, fmt.Errorf("adr: failed to read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("adr: failed to parse config %s: %w", path, err)
	}
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err()
	default:
	}
	return normalizeConfig(cfg)
}

func saveConfig(root string, cfg Config) error {
	return saveConfigContext(context.Background(), root, cfg)
}

func saveConfigContext(ctx context.Context, root string, cfg Config) error {
	path := configPath(root)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := shared.EnsureInDir(root, path); err != nil {
		return err
	}
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("adr: failed to marshal config: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("adr: failed to create adr root %s: %w", root, err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("adr: failed to write config %s: %w", path, err)
	}
	return nil
}
