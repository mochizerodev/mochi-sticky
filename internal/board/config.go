package board

import (
	"context"
	"fmt"
	"os"
	"strings"

	"mochi-sticky/internal/shared"

	"gopkg.in/yaml.v3"
)

// Column defines a single status column.
type Column struct {
	Key   string `yaml:"key"`
	Title string `yaml:"title"`
}

// BoardContext captures metadata that applies to an entire board.
type BoardContext struct {
	Scope   string   `yaml:"scope,omitempty" json:"scope,omitempty"`
	Owners  []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	Release string   `yaml:"release,omitempty" json:"release,omitempty"`
	Target  string   `yaml:"target,omitempty" json:"target,omitempty"`
	Notes   string   `yaml:"notes,omitempty" json:"notes,omitempty"`
}

// Config defines the sticky board configuration.
type Config struct {
	ConfigVersion int          `yaml:"config_version" json:"config_version"`
	NextID        int          `yaml:"next_id" json:"next_id"`
	Columns       []Column     `yaml:"columns" json:"columns"`
	Context       BoardContext `yaml:"context,omitempty" json:"context,omitempty"`
}

// DefaultConfig returns the default board configuration.
func DefaultConfig() Config {
	return Config{
		ConfigVersion: 1,
		NextID:        1,
		Columns: []Column{
			{Key: "todo", Title: "Todo"},
			{Key: "doing", Title: "Doing"},
			{Key: "done", Title: "Done"},
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

// ParseConfig reads a board config from YAML bytes.
func ParseConfig(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("board: failed to parse config: %w", err)
	}
	return normalizeConfig(cfg), nil
}

// SaveConfig writes a board config file to disk.
func (r *Repository) SaveConfig(cfg Config) error {
	return r.SaveConfigContext(context.Background(), cfg)
}

// SaveConfigContext writes a board config file to disk, honoring ctx cancellation.
func (r *Repository) SaveConfigContext(ctx context.Context, cfg Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.saveConfigContext(ctx, normalizeConfig(cfg))
}

// LoadConfig reads the config file from disk.
func (r *Repository) LoadConfig() (Config, error) {
	return r.LoadConfigContext(context.Background())
}

// LoadConfigContext reads the config file from disk, honoring ctx cancellation.
func (r *Repository) LoadConfigContext(ctx context.Context) (Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.loadConfigContext(ctx)
}

func (r *Repository) loadConfig() (Config, error) {
	return r.loadConfigContext(context.Background())
}

func (r *Repository) loadConfigContext(ctx context.Context) (Config, error) {
	if r.configPath == "" {
		return Config{}, fmt.Errorf("board: %w", shared.ErrInvalidPath)
	}
	configPath := r.configPath
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err()
	default:
	}
	if err := shared.EnsureInDir(r.boardDir, configPath); err != nil {
		return Config{}, err
	}
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err()
	default:
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("board: %w", ErrStoreNotInitialized)
		}
		return Config{}, fmt.Errorf("board: failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("board: failed to parse config: %w", err)
	}
	select {
	case <-ctx.Done():
		return Config{}, ctx.Err()
	default:
	}
	if cfg.NextID <= 0 {
		cfg.NextID = 1
	}
	return cfg, nil
}

func (r *Repository) saveConfig(cfg Config) error {
	return r.saveConfigContext(context.Background(), cfg)
}

func (r *Repository) saveConfigContext(ctx context.Context, cfg Config) error {
	if r.configPath == "" {
		return fmt.Errorf("board: %w", shared.ErrInvalidPath)
	}
	configPath := r.configPath
	if err := shared.EnsureInDir(r.boardDir, configPath); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("board: failed to marshal config: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("board: failed to write config file: %w", err)
	}
	return nil
}

// UpdateBoardContext stores the provided context block in the board config.
func (r *Repository) UpdateBoardContext(ctx BoardContext) error {
	return r.UpdateBoardContextContext(context.Background(), ctx)
}

// UpdateBoardContextContext stores the provided context block in the board config, honoring ctx cancellation.
func (r *Repository) UpdateBoardContextContext(ctx context.Context, boardCtx BoardContext) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, err := r.loadConfigContext(ctx)
	if err != nil {
		return err
	}
	cfg.Context = normalizeBoardContext(boardCtx)
	return r.saveConfigContext(ctx, cfg)
}

func normalizeBoardContext(ctx BoardContext) BoardContext {
	cleanOwners := make([]string, 0, len(ctx.Owners))
	for _, owner := range ctx.Owners {
		trimmed := strings.TrimSpace(owner)
		if trimmed == "" {
			continue
		}
		cleanOwners = append(cleanOwners, trimmed)
	}
	ctx.Owners = cleanOwners
	ctx.Scope = strings.TrimSpace(ctx.Scope)
	ctx.Release = strings.TrimSpace(ctx.Release)
	ctx.Target = strings.TrimSpace(ctx.Target)
	ctx.Notes = strings.TrimSpace(ctx.Notes)
	return ctx
}

func normalizeConfig(cfg Config) Config {
	if cfg.ConfigVersion <= 0 {
		cfg.ConfigVersion = 1
	}
	if cfg.NextID <= 0 {
		cfg.NextID = 1
	}
	if len(cfg.Columns) == 0 {
		cfg.Columns = DefaultConfig().Columns
		return cfg
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
		if title == "" {
			title = key
		}
		clean = append(clean, Column{
			Key:   key,
			Title: title,
		})
	}
	if len(clean) == 0 {
		cfg.Columns = DefaultConfig().Columns
	} else {
		cfg.Columns = clean
	}
	return cfg
}
