package board

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/shared"

	"gopkg.in/yaml.v3"
)

// Board represents a persistable board entry.
type Board struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	Archived bool   `yaml:"archived"`
	Created  Date   `yaml:"created"`
}

// BoardRegistry stores the list of boards and the active board ID for the storage.
type BoardRegistry struct {
	Active string  `yaml:"active"`
	Boards []Board `yaml:"boards"`
}

// LoadBoardRegistry reads the board registry from disk using the owning repository.
func (r *Repository) LoadBoardRegistry() (BoardRegistry, error) {
	return r.LoadBoardRegistryContext(context.Background())
}

// LoadBoardRegistryContext reads the board registry from disk, honoring ctx cancellation.
func (r *Repository) LoadBoardRegistryContext(ctx context.Context) (BoardRegistry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.loadBoardRegistryContext(ctx)
}

func (r *Repository) loadBoardRegistry() (BoardRegistry, error) {
	return r.loadBoardRegistryContext(context.Background())
}

func (r *Repository) loadBoardRegistryContext(ctx context.Context) (BoardRegistry, error) {
	registryPath := r.registryPath
	select {
	case <-ctx.Done():
		return BoardRegistry{}, ctx.Err()
	default:
	}
	if err := shared.EnsureInDir(r.stickyDir, registryPath); err != nil {
		return BoardRegistry{}, err
	}
	select {
	case <-ctx.Done():
		return BoardRegistry{}, ctx.Err()
	default:
	}
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return BoardRegistry{}, fmt.Errorf("board: %w", ErrStoreNotInitialized)
		}
		return BoardRegistry{}, fmt.Errorf("board: failed to read board registry: %w", err)
	}

	var registry BoardRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return BoardRegistry{}, fmt.Errorf("board: failed to parse board registry: %w", err)
	}
	select {
	case <-ctx.Done():
		return BoardRegistry{}, ctx.Err()
	default:
	}
	return registry, nil
}

// saveBoardRegistryContext persists the registry YAML, honoring ctx cancellation.
func (r *Repository) saveBoardRegistryContext(ctx context.Context, registry BoardRegistry) error {
	registryPath := r.registryPath
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := shared.EnsureInDir(r.stickyDir, registryPath); err != nil {
		return err
	}
	data, err := yaml.Marshal(registry)
	if err != nil {
		return fmt.Errorf("board: failed to marshal board registry: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(registryPath, data, 0o644); err != nil {
		return fmt.Errorf("board: failed to write board registry: %w", err)
	}
	return nil
}

// findBoard locates a board in the provided registry.
func findBoard(registry BoardRegistry, boardID string) (Board, error) {
	target := strings.TrimSpace(boardID)
	for _, board := range registry.Boards {
		if board.ID == target {
			return board, nil
		}
	}
	return Board{}, fmt.Errorf("board: %w", ErrBoardNotFound)
}

// validateBoardID enforces the allowed board ID format.
func validateBoardID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("board: %w", ErrInvalidBoardID)
	}
	if strings.ContainsAny(id, `/\`) || strings.Contains(id, string(filepath.Separator)) {
		return fmt.Errorf("board: %w", ErrInvalidBoardID)
	}
	return nil
}

// generateBoardID slugifies the provided name and avoids collisions with existing IDs.
func generateBoardID(name string, boards []Board) string {
	base := slugify(name)
	if base == "" {
		base = "board"
	}
	id := base
	for i := 1; boardIDExists(id, boards); i++ {
		id = fmt.Sprintf("%s-%d", base, i)
	}
	return id
}

// boardIDExists returns true if any board already uses the given ID.
func boardIDExists(id string, boards []Board) bool {
	for _, board := range boards {
		if board.ID == id {
			return true
		}
	}
	return false
}

// slugify creates a filesystem-safe slug used for board IDs and paths.
func slugify(value string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			prevDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	return slug
}

// firstActiveBoard picks the next active board ID while skipping skipID.
func firstActiveBoard(boards []Board, skipID string) string {
	for _, board := range boards {
		if board.ID == skipID {
			continue
		}
		if board.Archived {
			continue
		}
		return board.ID
	}
	for _, board := range boards {
		if board.ID == skipID {
			continue
		}
		return board.ID
	}
	return ""
}
