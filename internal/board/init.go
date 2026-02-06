package board

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/shared"
)

// InitStore scaffolds the `.sticky` storage layout, board registry, and default config.
func (r *Repository) InitStore() error {
	return r.InitStoreContext(context.Background())
}

// InitStoreContext scaffolds the `.sticky` storage layout, honoring ctx cancellation.
func (r *Repository) InitStoreContext(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := r.migrateLegacyLayoutLockedContext(ctx); err != nil {
		return err
	}
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := r.ensureBoardRegistryContext(ctx); err != nil {
		return err
	}
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(r.tasksDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create tasks directory: %w", err)
	}
	if err := checkCtx(ctx); err != nil {
		return err
	}
	wikiDir := filepath.Join(r.stickyDir, "wiki")
	if err := os.MkdirAll(wikiDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create wiki directory: %w", err)
	}
	if err := checkCtx(ctx); err != nil {
		return err
	}
	adrDir := filepath.Join(r.stickyDir, "adrs")
	adrRepo, err := adr.NewRepository(adrDir)
	if err != nil {
		return fmt.Errorf("board: failed to configure adr store: %w", err)
	}
	if err := adrRepo.InitStoreContext(ctx); err != nil {
		return fmt.Errorf("board: failed to initialize adr store: %w", err)
	}

	if err := r.writeDefaultConfigContext(ctx); err != nil {
		return err
	}
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(r.archiveTasks, 0o755); err != nil {
		return fmt.Errorf("board: failed to create archive tasks directory: %w", err)
	}
	if err := r.ensureGitignoreContext(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repository) writeDefaultConfigContext(ctx context.Context) error {
	if r.configPath == "" {
		return fmt.Errorf("board: %w", shared.ErrInvalidPath)
	}
	configPath := r.configPath
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := shared.EnsureInDir(r.boardDir, configPath); err != nil {
		return err
	}
	if _, err := os.Stat(configPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("board: failed to stat config file: %w", err)
	}

	data, err := RenderConfig()
	if err != nil {
		return fmt.Errorf("board: failed to render config: %w", err)
	}
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("board: failed to write config file: %w", err)
	}
	return nil
}

func (r *Repository) ensureGitignoreContext(ctx context.Context) error {
	gitignorePath := filepath.Join(r.stickyDir, ".gitignore")
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := shared.EnsureInDir(r.stickyDir, gitignorePath); err != nil {
		return err
	}

	desiredLines := []string{".sticky/debug.log", "debug.log"}
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		content := strings.Join(desiredLines, "\n") + "\n"
		if err := checkCtx(ctx); err != nil {
			return err
		}
		if err := os.WriteFile(gitignorePath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("board: failed to write .gitignore: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("board: failed to stat .gitignore: %w", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("board: failed to read .gitignore: %w", err)
	}

	missing := make([]string, 0, len(desiredLines))
	for _, line := range desiredLines {
		if !hasLine(data, line) {
			missing = append(missing, line)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	content := strings.TrimRight(string(data), "\n") + "\n" + strings.Join(missing, "\n") + "\n"
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := os.WriteFile(gitignorePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("board: failed to update .gitignore: %w", err)
	}
	return nil
}

func hasLine(data []byte, line string) bool {
	for _, existing := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(existing) == line {
			return true
		}
	}
	return false
}

func (r *Repository) ensureBoardRegistryContext(ctx context.Context) error {
	registryPath := r.registryPath
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := shared.EnsureInDir(r.stickyDir, registryPath); err != nil {
		return err
	}
	if _, err := os.Stat(registryPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("board: failed to stat board registry: %w", err)
	}

	boardsDir := filepath.Dir(registryPath)
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(boardsDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create boards directory: %w", err)
	}

	boardID := "default"
	boardPath := filepath.Join("boards", boardID)
	registry := BoardRegistry{
		Active: boardID,
		Boards: []Board{
			{
				ID:       boardID,
				Name:     "Default",
				Path:     boardPath,
				Archived: false,
				Created:  Date{Time: r.now()},
			},
		},
	}

	if err := r.saveBoardRegistryContext(ctx, registry); err != nil {
		return err
	}

	boardDir := filepath.Join(r.stickyDir, boardPath)
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := shared.EnsureInDir(r.stickyDir, boardDir); err != nil {
		return err
	}
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(boardDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create default board directory: %w", err)
	}
	return nil
}

func (r *Repository) migrateLegacyLayoutLockedContext(ctx context.Context) error {
	if !r.legacyLayout {
		return nil
	}

	legacyTasks := filepath.Join(r.stickyDir, "tasks")
	legacyConfig := filepath.Join(r.stickyDir, "config.yaml")

	boardID := "default"
	boardPath := filepath.Join("boards", boardID)
	boardDir := filepath.Join(r.stickyDir, boardPath)
	tasksDir := filepath.Join(boardDir, "tasks")
	configPath := filepath.Join(boardDir, "config.yaml")

	if err := checkCtx(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(boardDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create board directory: %w", err)
	}

	if err := checkCtx(ctx); err != nil {
		return err
	}
	if exists, err := shared.PathExists(legacyTasks); err != nil {
		return err
	} else if exists {
		if targetExists, err := shared.PathExists(tasksDir); err != nil {
			return err
		} else if targetExists {
			return fmt.Errorf("board: failed to migrate legacy tasks: %w", shared.ErrInvalidPath)
		}
		if err := os.Rename(legacyTasks, tasksDir); err != nil {
			return fmt.Errorf("board: failed to move legacy tasks: %w", err)
		}
	}

	if err := checkCtx(ctx); err != nil {
		return err
	}
	if exists, err := shared.PathExists(legacyConfig); err != nil {
		return err
	} else if exists {
		if targetExists, err := shared.PathExists(configPath); err != nil {
			return err
		} else if targetExists {
			return fmt.Errorf("board: failed to migrate legacy config: %w", shared.ErrInvalidPath)
		}
		if err := os.Rename(legacyConfig, configPath); err != nil {
			return fmt.Errorf("board: failed to move legacy config: %w", err)
		}
	}

	registry := BoardRegistry{
		Active: boardID,
		Boards: []Board{
			{
				ID:       boardID,
				Name:     "Default",
				Path:     boardPath,
				Archived: false,
				Created:  Date{Time: r.now()},
			},
		},
	}
	if err := r.saveBoardRegistryContext(ctx, registry); err != nil {
		return err
	}

	r.applyBoard(Board{ID: boardID, Path: boardPath})
	r.legacyLayout = false
	return nil
}

func checkCtx(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
