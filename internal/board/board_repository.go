package board

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mochi-sticky/internal/shared"
	"mochi-sticky/internal/storage"

	"gopkg.in/yaml.v3"
)

// BoardRepository manages the board registry and board defaults.
type BoardRepository struct {
	mu           sync.RWMutex
	baseDir      string
	stickyDir    string
	registryPath string
	now          func() time.Time
}

// NewBoardRepository creates a board repository rooted at baseDir.
func NewBoardRepository(baseDir string) (*BoardRepository, error) {
	return NewBoardRepositoryWithStorage(baseDir, "")
}

// NewBoardRepositoryWithStorage creates a board repository rooted at baseDir with a custom storage root.
func NewBoardRepositoryWithStorage(baseDir, storageRoot string) (*BoardRepository, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("board: base directory is required: %w", shared.ErrInvalidPath)
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("board: failed to resolve base directory: %w", err)
	}
	stickyDir := storageRoot
	if strings.TrimSpace(stickyDir) == "" {
		stickyDir = filepath.Join(absBase, ".sticky")
	} else if !filepath.IsAbs(stickyDir) {
		stickyDir = filepath.Join(absBase, stickyDir)
	}
	stickyDir, err = filepath.Abs(stickyDir)
	if err != nil {
		return nil, fmt.Errorf("board: failed to resolve storage root: %w", err)
	}

	cfg, err := storage.LoadConfigFromRoot(stickyDir)
	if err != nil {
		return nil, err
	}
	paths, err := storage.ResolveConfigPaths(absBase, stickyDir, cfg)
	if err != nil {
		return nil, err
	}
	registryPath := paths.Boards
	if strings.TrimSpace(cfg.Paths.Boards) == "" {
		legacyRegistry := filepath.Join(stickyDir, "boards.yaml")
		if _, err := os.Stat(registryPath); err != nil && os.IsNotExist(err) {
			if _, legacyErr := os.Stat(legacyRegistry); legacyErr == nil {
				registryPath = legacyRegistry
			}
		}
	}

	return &BoardRepository{
		baseDir:      absBase,
		stickyDir:    stickyDir,
		registryPath: registryPath,
		now:          time.Now,
	}, nil
}

// LoadRegistry reads the board registry from disk.
func (b *BoardRepository) LoadRegistry() (BoardRegistry, error) {
	return b.LoadRegistryContext(context.Background())
}

// LoadRegistryContext reads the board registry from disk, honoring ctx cancellation.
func (b *BoardRepository) LoadRegistryContext(ctx context.Context) (BoardRegistry, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.loadRegistryContext(ctx)
}

func (b *BoardRepository) loadRegistryContext(ctx context.Context) (BoardRegistry, error) {
	registryPath := b.registryPath
	select {
	case <-ctx.Done():
		return BoardRegistry{}, ctx.Err()
	default:
	}
	if err := shared.EnsureInDir(b.stickyDir, registryPath); err != nil {
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

func (b *BoardRepository) saveRegistryContext(ctx context.Context, registry BoardRegistry) error {
	registryPath := b.registryPath
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := shared.EnsureInDir(b.stickyDir, registryPath); err != nil {
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

// ListBoards returns all boards and the active board id.
// ListBoards returns all configured boards and the currently active board ID.
func (b *BoardRepository) ListBoards() ([]Board, string, error) {
	return b.ListBoardsContext(context.Background())
}

// ListBoardsContext returns all configured boards and the currently active board ID, honoring ctx cancellation.
func (b *BoardRepository) ListBoardsContext(ctx context.Context) ([]Board, string, error) {
	registry, err := b.LoadRegistryContext(ctx)
	if err != nil {
		return nil, "", err
	}
	return registry.Boards, registry.Active, nil
}

// CreateBoard registers a new board and initializes its storage.
// CreateBoard adds a new board entry to the registry and initializes its directories.
func (b *BoardRepository) CreateBoard(name string) (Board, error) {
	return b.CreateBoardContext(context.Background(), name)
}

// CreateBoardContext adds a new board entry to the registry and initializes its directories.
func (b *BoardRepository) CreateBoardContext(ctx context.Context, name string) (Board, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-ctx.Done():
		return Board{}, ctx.Err()
	default:
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return Board{}, fmt.Errorf("board: %w", ErrInvalidTitle)
	}
	if err := b.ensureRegistryContext(ctx); err != nil {
		return Board{}, err
	}
	registry, err := b.loadRegistryContext(ctx)
	if err != nil {
		return Board{}, err
	}

	id := generateBoardID(trimmed, registry.Boards)
	board := Board{
		ID:       id,
		Name:     trimmed,
		Path:     filepath.Join("boards", id),
		Archived: false,
		Created:  Date{Time: b.now()},
	}
	registry.Boards = append(registry.Boards, board)
	if registry.Active == "" {
		registry.Active = board.ID
	}
	if err := b.saveRegistryContext(ctx, registry); err != nil {
		return Board{}, err
	}
	if err := b.initBoardLockedContext(ctx, board); err != nil {
		return Board{}, err
	}
	return board, nil
}

// RenameBoard updates a board's name.
// RenameBoard updates the name of an existing board.
func (b *BoardRepository) RenameBoard(boardID, name string) (Board, error) {
	return b.RenameBoardContext(context.Background(), boardID, name)
}

// RenameBoardContext updates the name of an existing board, honoring ctx cancellation.
func (b *BoardRepository) RenameBoardContext(ctx context.Context, boardID, name string) (Board, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-ctx.Done():
		return Board{}, ctx.Err()
	default:
	}
	if err := validateBoardID(boardID); err != nil {
		return Board{}, err
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return Board{}, fmt.Errorf("board: %w", ErrInvalidTitle)
	}
	registry, err := b.loadRegistryContext(ctx)
	if err != nil {
		return Board{}, err
	}

	updated := false
	var result Board
	for i, board := range registry.Boards {
		if board.ID != boardID {
			continue
		}
		board.Name = trimmed
		registry.Boards[i] = board
		result = board
		updated = true
		break
	}
	if !updated {
		return Board{}, fmt.Errorf("board: %w", ErrBoardNotFound)
	}
	if err := b.saveRegistryContext(ctx, registry); err != nil {
		return Board{}, err
	}
	return result, nil
}

// SetActiveBoard switches the active board.
// SetActiveBoard marks a board as the active board in the registry.
func (b *BoardRepository) SetActiveBoard(boardID string) error {
	return b.SetActiveBoardContext(context.Background(), boardID)
}

// SetActiveBoardContext marks a board as the active board in the registry, honoring ctx cancellation.
func (b *BoardRepository) SetActiveBoardContext(ctx context.Context, boardID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := validateBoardID(boardID); err != nil {
		return err
	}
	registry, err := b.loadRegistryContext(ctx)
	if err != nil {
		return err
	}
	if _, err := findBoard(registry, boardID); err != nil {
		return err
	}
	registry.Active = boardID
	return b.saveRegistryContext(ctx, registry)
}

// ArchiveBoard sets the archived flag on a board.
// ArchiveBoard archives the requested board so it can no longer be used.
func (b *BoardRepository) ArchiveBoard(boardID string) (Board, error) {
	return b.ArchiveBoardContext(context.Background(), boardID)
}

// ArchiveBoardContext archives the requested board so it can no longer be used, honoring ctx cancellation.
func (b *BoardRepository) ArchiveBoardContext(ctx context.Context, boardID string) (Board, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-ctx.Done():
		return Board{}, ctx.Err()
	default:
	}
	if err := validateBoardID(boardID); err != nil {
		return Board{}, err
	}
	registry, err := b.loadRegistryContext(ctx)
	if err != nil {
		return Board{}, err
	}
	updated := false
	var result Board
	for i, board := range registry.Boards {
		if board.ID != boardID {
			continue
		}
		board.Archived = true
		registry.Boards[i] = board
		result = board
		updated = true
		break
	}
	if !updated {
		return Board{}, fmt.Errorf("board: %w", ErrBoardNotFound)
	}
	if registry.Active == boardID {
		if next := firstActiveBoard(registry.Boards, boardID); next != "" {
			registry.Active = next
		}
	}
	if err := b.saveRegistryContext(ctx, registry); err != nil {
		return Board{}, err
	}
	return result, nil
}

// DeleteBoard removes a board and deletes its data directory.
// DeleteBoard removes a board entry from the registry (board must already be archived).
func (b *BoardRepository) DeleteBoard(boardID string) error {
	return b.DeleteBoardContext(context.Background(), boardID)
}

// DeleteBoardContext removes a board entry from the registry and deletes its data directory,
// honoring ctx cancellation.
func (b *BoardRepository) DeleteBoardContext(ctx context.Context, boardID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := validateBoardID(boardID); err != nil {
		return err
	}
	registry, err := b.loadRegistryContext(ctx)
	if err != nil {
		return err
	}
	if len(registry.Boards) <= 1 {
		return fmt.Errorf("board: %w", ErrBoardDeleteForbidden)
	}

	var target Board
	filtered := make([]Board, 0, len(registry.Boards)-1)
	for _, board := range registry.Boards {
		if board.ID == boardID {
			target = board
			continue
		}
		filtered = append(filtered, board)
	}
	if target.ID == "" {
		return fmt.Errorf("board: %w", ErrBoardNotFound)
	}

	registry.Boards = filtered
	if registry.Active == boardID {
		registry.Active = firstActiveBoard(filtered, "")
		if registry.Active == "" {
			return fmt.Errorf("board: %w", ErrBoardDeleteForbidden)
		}
	}
	if err := b.saveRegistryContext(ctx, registry); err != nil {
		return err
	}

	boardDir := b.boardDirFor(target)
	if err := shared.EnsureInDir(b.stickyDir, boardDir); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.RemoveAll(boardDir); err != nil {
		return fmt.Errorf("board: failed to delete board data: %w", err)
	}
	return nil
}

// ResolveBoardPaths returns the board and its resolved directories.
// ResolveBoardPaths returns the board record plus its root, config, tasks, and archive directories.
func (b *BoardRepository) ResolveBoardPaths(boardID string) (Board, string, string, string, error) {
	registry, err := b.LoadRegistry()
	if err != nil {
		return Board{}, "", "", "", err
	}
	targetID := strings.TrimSpace(boardID)
	if targetID == "" {
		targetID = registry.Active
	}
	board, err := findBoard(registry, targetID)
	if err != nil {
		return Board{}, "", "", "", err
	}
	boardDir := b.boardDirFor(board)
	tasksDir := filepath.Join(boardDir, "tasks")
	configPath := filepath.Join(boardDir, "config.yaml")
	return board, boardDir, tasksDir, configPath, nil
}

func (b *BoardRepository) ensureRegistryContext(ctx context.Context) error {
	registryPath := b.registryPath
	if err := shared.EnsureInDir(b.stickyDir, registryPath); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if _, err := os.Stat(registryPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("board: failed to stat board registry: %w", err)
	}

	boardsDir := filepath.Dir(registryPath)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
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
				Created:  Date{Time: b.now()},
			},
		},
	}

	if err := b.saveRegistryContext(ctx, registry); err != nil {
		return err
	}

	boardDir := filepath.Join(b.stickyDir, boardPath)
	if err := shared.EnsureInDir(b.stickyDir, boardDir); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(boardDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create default board directory: %w", err)
	}
	return nil
}

func (b *BoardRepository) initBoardLockedContext(ctx context.Context, board Board) error {
	if err := validateBoardID(board.ID); err != nil {
		return err
	}
	if err := b.ensureRegistryContext(ctx); err != nil {
		return err
	}

	boardDir := b.boardDirFor(board)
	if err := shared.EnsureInDir(b.stickyDir, boardDir); err != nil {
		return err
	}
	tasksDir := filepath.Join(boardDir, "tasks")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		return fmt.Errorf("board: failed to create board tasks directory: %w", err)
	}
	archiveTasks := filepath.Join(boardDir, "archive", "tasks")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(archiveTasks, 0o755); err != nil {
		return fmt.Errorf("board: failed to create board archive tasks directory: %w", err)
	}

	configPath := filepath.Join(boardDir, "config.yaml")
	if err := shared.EnsureInDir(boardDir, configPath); err != nil {
		return err
	}
	if _, err := os.Stat(configPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("board: failed to stat board config: %w", err)
	}

	data, err := RenderConfig()
	if err != nil {
		return fmt.Errorf("board: failed to render board config: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("board: failed to write board config: %w", err)
	}
	return nil
}

func (b *BoardRepository) boardDirFor(board Board) string {
	boardPath := strings.TrimSpace(board.Path)
	if boardPath == "" {
		boardPath = filepath.Join("boards", board.ID)
	}
	return filepath.Join(b.stickyDir, boardPath)
}
