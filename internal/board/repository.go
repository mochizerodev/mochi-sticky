package board

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mochi-sticky/internal/shared"
	"mochi-sticky/internal/storage"
)

// Repository manages task files on disk.
type Repository struct {
	mu           sync.RWMutex
	baseDir      string
	stickyDir    string
	boardID      string
	boardName    string
	boardDir     string
	tasksDir     string
	archiveDir   string
	archiveTasks string
	registryPath string
	configPath   string
	legacyLayout bool
	parser       *Parser
	now          func() time.Time
}

// BaseDir returns the repository base directory.
func (r *Repository) BaseDir() string {
	return r.baseDir
}

// StorageRoot returns the root directory used for all mochi-sticky data (default: .sticky).
func (r *Repository) StorageRoot() string {
	return r.stickyDir
}

// BoardID returns the active board ID for this repository.
func (r *Repository) BoardID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.boardID
}

// NewRepository creates a repository rooted at baseDir using the active board.
func NewRepository(baseDir string) (*Repository, error) {
	return NewRepositoryForBoard(baseDir, "")
}

// NewRepositoryWithStorage creates a repository rooted at baseDir with a custom storage root.
func NewRepositoryWithStorage(baseDir, storageRoot string) (*Repository, error) {
	return NewRepositoryForBoardWithStorage(baseDir, "", storageRoot)
}

// NewRepositoryForBoard creates a repository rooted at baseDir for a board.
func NewRepositoryForBoard(baseDir, boardID string) (*Repository, error) {
	return NewRepositoryForBoardWithStorage(baseDir, boardID, "")
}

// NewRepositoryForBoardWithStorage creates a repository rooted at baseDir with a custom storage root.
func NewRepositoryForBoardWithStorage(baseDir, boardID, storageRoot string) (*Repository, error) {
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

	repo := &Repository{
		baseDir:      absBase,
		stickyDir:    stickyDir,
		registryPath: registryPath,
		parser:       &Parser{},
		now:          time.Now,
	}
	if err := repo.selectBoard(boardID); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *Repository) selectBoard(boardID string) error {
	registryPath := r.registryPath
	if _, err := os.Stat(registryPath); err == nil {
		registry, err := r.loadBoardRegistry()
		if err != nil {
			return err
		}
		targetID := strings.TrimSpace(boardID)
		if targetID == "" {
			targetID = registry.Active
		}
		board, err := findBoard(registry, targetID)
		if err != nil {
			return err
		}
		r.applyBoard(board)
		if err := shared.EnsureInDir(r.stickyDir, r.boardDir); err != nil {
			return err
		}
		r.legacyLayout = false
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("board: failed to stat board registry: %w", err)
	}

	if strings.TrimSpace(boardID) != "" {
		return fmt.Errorf("board: %w", ErrBoardNotFound)
	}

	if legacyLayoutExists(r.stickyDir) {
		r.boardID = "default"
		r.boardName = "default"
		r.boardDir = r.stickyDir
		r.tasksDir = filepath.Join(r.stickyDir, "tasks")
		r.archiveDir = filepath.Join(r.stickyDir, "archive")
		r.archiveTasks = filepath.Join(r.archiveDir, "tasks")
		r.configPath = filepath.Join(r.stickyDir, "config.yaml")
		r.legacyLayout = true
		return nil
	}

	r.boardID = "default"
	r.boardName = "default"
	r.boardDir = filepath.Join(r.stickyDir, "boards", "default")
	r.tasksDir = filepath.Join(r.boardDir, "tasks")
	r.archiveDir = filepath.Join(r.boardDir, "archive")
	r.archiveTasks = filepath.Join(r.archiveDir, "tasks")
	r.configPath = filepath.Join(r.boardDir, "config.yaml")
	r.legacyLayout = false
	return nil
}

func (r *Repository) applyBoard(board Board) {
	r.boardID = board.ID
	boardPath := strings.TrimSpace(board.Path)
	if boardPath == "" {
		boardPath = filepath.Join("boards", board.ID)
	}
	r.boardDir = filepath.Join(r.stickyDir, boardPath)
	r.tasksDir = filepath.Join(r.boardDir, "tasks")
	r.archiveDir = filepath.Join(r.boardDir, "archive")
	r.archiveTasks = filepath.Join(r.archiveDir, "tasks")
	r.configPath = filepath.Join(r.boardDir, "config.yaml")
	r.boardName = strings.TrimSpace(board.Name)
	if r.boardName == "" {
		r.boardName = board.ID
	}
}

// GetAllTasks returns all tasks from the tasks directory.
// GetAllTasks returns all task files currently stored for the repo's active board.
func (r *Repository) GetAllTasks() ([]Task, error) {
	return r.GetAllTasksContext(context.Background())
}

// GetAllTasksContext returns all task files currently stored for the repo's active board,
// honoring ctx cancellation.
func (r *Repository) GetAllTasksContext(ctx context.Context) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := ensureDirExists(r.tasksDir); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(r.tasksDir)
	if err != nil {
		return nil, fmt.Errorf("board: failed to read tasks directory: %w", err)
	}

	var tasks []Task
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(r.tasksDir, entry.Name())
		if err := shared.EnsureInDir(r.tasksDir, path); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("board: failed to read task file %s: %w", path, err)
		}
		task, err := r.parser.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("board: failed to parse task file %s: %w", path, err)
		}
		task.FilePath = path
		r.attachBoardInfo(&task)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// ListReadyTasks returns tasks whose dependencies are all satisfied.
// ListReadyTasks returns tasks whose dependencies have all been satisfied.
func (r *Repository) ListReadyTasks() ([]Task, error) {
	return r.ListReadyTasksContext(context.Background())
}

// ListReadyTasksContext returns tasks whose dependencies have all been satisfied, honoring ctx cancellation.
func (r *Repository) ListReadyTasksContext(ctx context.Context) ([]Task, error) {
	all, err := r.GetAllTasksContext(ctx)
	if err != nil {
		return nil, err
	}
	index := make(map[string]Task, len(all))
	for _, t := range all {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		index[t.ID] = t
	}
	var ready []Task
	for _, t := range all {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		ok, _ := IsReady(t, index)
		if ok {
			ready = append(ready, t)
		}
	}
	return ready, nil
}

// CreateTask creates a new task file in the tasks directory.
// CreateTask writes a new task file, assigning IDs when needed and updating metadata.
func (r *Repository) CreateTask(task Task) (Task, error) {
	return r.CreateTaskContext(context.Background(), task)
}

// CreateTaskContext writes a new task file, honoring ctx cancellation.
func (r *Repository) CreateTaskContext(ctx context.Context, task Task) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	if err := os.MkdirAll(r.tasksDir, 0o755); err != nil {
		return Task{}, fmt.Errorf("board: failed to create tasks directory: %w", err)
	}

	if task.ID == "" {
		config, err := r.loadConfig()
		if err != nil {
			return Task{}, err
		}
		id := formatSequentialID(config.NextID)
		config.NextID++
		if err := r.saveConfig(config); err != nil {
			return Task{}, fmt.Errorf("board: failed to update config: %w", err)
		}
		task.ID = id
	}
	if task.UID == "" {
		uid, err := newID()
		if err != nil {
			return Task{}, fmt.Errorf("board: failed to generate task uid: %w", err)
		}
		task.UID = uid
	}
	if err := validateID(task.ID); err != nil {
		return Task{}, err
	}
	if task.Created.IsZero() {
		task.Created = Date{Time: r.now()}
	}
	priority, err := normalizePriority(task.Priority)
	if err != nil {
		return Task{}, err
	}
	task.Priority = priority

	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	content, err := r.parser.Render(task)
	if err != nil {
		return Task{}, fmt.Errorf("board: failed to render task %s: %w", task.ID, err)
	}

	filePath := filepath.Join(r.tasksDir, task.ID+".md")
	if err := shared.EnsureInDir(r.tasksDir, filePath); err != nil {
		return Task{}, err
	}
	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return Task{}, fmt.Errorf("board: failed to write task file %s: %w", filePath, err)
	}

	task.FilePath = filePath
	return task, nil
}

func formatSequentialID(value int) string {
	return fmt.Sprintf("T-%06d", value)
}

// UpdateTaskStatus updates the status of a task by ID.
// UpdateTaskStatus changes the status of the task with the provided ID.
func (r *Repository) UpdateTaskStatus(id string, status string) error {
	return r.UpdateTaskStatusContext(context.Background(), id, status)
}

// UpdateTaskStatusContext changes the status of the task with the provided ID, honoring ctx cancellation.
func (r *Repository) UpdateTaskStatusContext(ctx context.Context, id string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := validateID(id); err != nil {
		return err
	}
	if err := ensureDirExists(r.tasksDir); err != nil {
		return err
	}

	entries, err := os.ReadDir(r.tasksDir)
	if err != nil {
		return fmt.Errorf("board: failed to read tasks directory: %w", err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(r.tasksDir, entry.Name())
		if err := shared.EnsureInDir(r.tasksDir, path); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("board: failed to read task file %s: %w", path, err)
		}
		task, err := r.parser.Parse(data)
		if err != nil {
			return fmt.Errorf("board: failed to parse task file %s: %w", path, err)
		}
		if task.ID != id {
			continue
		}
		task.Status = status
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		content, err := r.parser.Render(task)
		if err != nil {
			return fmt.Errorf("board: failed to render task %s: %w", task.ID, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("board: failed to write task file %s: %w", path, err)
		}
		return nil
	}

	return fmt.Errorf("board: %w", ErrTaskNotFound)
}

// UpdateTaskDependencies sets the dependency list for a task and validates cycles.
// UpdateTaskDependencies overwrites the dependency list for the given task.
func (r *Repository) UpdateTaskDependencies(id string, deps []string) error {
	return r.UpdateTaskDependenciesContext(context.Background(), id, deps)
}

// UpdateTaskDependenciesContext overwrites the dependency list for the given task, honoring ctx cancellation.
func (r *Repository) UpdateTaskDependenciesContext(ctx context.Context, id string, deps []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := validateID(id); err != nil {
		return err
	}
	for _, dep := range deps {
		if err := validateID(dep); err != nil {
			return fmt.Errorf("board: dependency %q is invalid: %w", dep, ErrInvalidDependency)
		}
	}
	if err := ensureDirExists(r.tasksDir); err != nil {
		return err
	}

	tasks, err := r.readTasksFromDirContext(ctx, r.tasksDir)
	if err != nil {
		return err
	}

	found := false
	for i := range tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if tasks[i].ID == id {
			tasks[i].DependsOn = normalizeIDs(deps)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("board: %w", ErrTaskNotFound)
	}
	if err := ValidateNoCycles(tasks); err != nil {
		return err
	}

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if task.ID != id {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		content, err := r.parser.Render(task)
		if err != nil {
			return fmt.Errorf("board: failed to render task %s: %w", task.ID, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := os.WriteFile(task.FilePath, content, 0o644); err != nil {
			return fmt.Errorf("board: failed to write task file %s: %w", task.FilePath, err)
		}
		return nil
	}

	return fmt.Errorf("board: %w", ErrTaskNotFound)
}

// GetTaskByID returns a task by ID.
// GetTaskByID loads a task by ID, searching the active tasks directory.
func (r *Repository) GetTaskByID(id string) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := validateID(id); err != nil {
		return Task{}, err
	}
	if err := ensureDirExists(r.tasksDir); err != nil {
		return Task{}, err
	}

	entries, err := os.ReadDir(r.tasksDir)
	if err != nil {
		return Task{}, fmt.Errorf("board: failed to read tasks directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(r.tasksDir, entry.Name())
		if err := shared.EnsureInDir(r.tasksDir, path); err != nil {
			return Task{}, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return Task{}, fmt.Errorf("board: failed to read task file %s: %w", path, err)
		}
		task, err := r.parser.Parse(data)
		if err != nil {
			return Task{}, fmt.Errorf("board: failed to parse task file %s: %w", path, err)
		}
		if task.ID != id {
			continue
		}
		task.FilePath = path
		r.attachBoardInfo(&task)
		return task, nil
	}

	return Task{}, fmt.Errorf("board: %w", ErrTaskNotFound)
}

func ensureDirExists(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("board: %w", ErrStoreNotInitialized)
		}
		return fmt.Errorf("board: failed to stat tasks directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("board: %w", shared.ErrInvalidPath)
	}
	return nil
}

func (r *Repository) attachBoardInfo(task *Task) {
	task.BoardID = r.boardID
	task.BoardName = r.boardName
	if strings.TrimSpace(task.BoardName) == "" {
		task.BoardName = r.boardID
	}
}

func validateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("board: %w", ErrInvalidID)
	}
	if strings.Contains(id, string(filepath.Separator)) {
		return fmt.Errorf("board: %w", ErrInvalidID)
	}
	return nil
}

func newID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(raw[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexStr[0:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:32]), nil
}

func legacyLayoutExists(stickyDir string) bool {
	tasksPath := filepath.Join(stickyDir, "tasks")
	if info, err := os.Stat(tasksPath); err == nil && info.IsDir() {
		return true
	}
	configPath := filepath.Join(stickyDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return true
	}
	return false
}
