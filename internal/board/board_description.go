package board

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/shared"
)

const boardDescriptionFilename = "board.md"

// BoardDescriptionPath returns the path where the board description markdown is stored.
func (r *Repository) BoardDescriptionPath() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.boardDescriptionPath()
}

// LoadBoardDescription reads the current board's description markdown.
func (r *Repository) LoadBoardDescription() (string, error) {
	return r.LoadBoardDescriptionContext(context.Background())
}

// LoadBoardDescriptionContext reads the current board's description markdown, honoring ctx cancellation.
func (r *Repository) LoadBoardDescriptionContext(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	path, err := r.boardDescriptionPath()
	if err != nil {
		return "", err
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("board: failed to read board description: %w", err)
	}
	return string(data), nil
}

// UpdateBoardDescription writes the current board's description markdown.
//
// Passing an empty/whitespace-only description removes the board.md file when present.
func (r *Repository) UpdateBoardDescription(description string) error {
	return r.UpdateBoardDescriptionContext(context.Background(), description)
}

// UpdateBoardDescriptionContext writes the current board's description markdown, honoring ctx cancellation.
//
// Passing an empty/whitespace-only description removes the board.md file when present.
func (r *Repository) UpdateBoardDescriptionContext(ctx context.Context, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path, err := r.boardDescriptionPath()
	if err != nil {
		return err
	}

	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("board: failed to remove board description: %w", err)
		}
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(r.boardDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create board directory: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(path, []byte(description), 0o644); err != nil {
		return fmt.Errorf("board: failed to write board description: %w", err)
	}
	return nil
}

func (r *Repository) boardDescriptionPath() (string, error) {
	if strings.TrimSpace(r.boardDir) == "" {
		return "", fmt.Errorf("board: %w", shared.ErrInvalidPath)
	}
	path := filepath.Join(r.boardDir, boardDescriptionFilename)
	if err := shared.EnsureInDir(r.boardDir, path); err != nil {
		return "", err
	}
	return path, nil
}
