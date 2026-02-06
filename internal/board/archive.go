package board

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mochi-sticky/internal/shared"
)

// ArchiveTask moves a task to the archive directory.
func (r *Repository) ArchiveTask(id string) (Task, error) {
	return r.ArchiveTaskContext(context.Background(), id)
}

// ArchiveTaskContext moves a task to the archive directory, honoring ctx cancellation.
func (r *Repository) ArchiveTaskContext(ctx context.Context, id string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	if err := validateID(id); err != nil {
		return Task{}, err
	}
	if err := ensureDirExists(r.tasksDir); err != nil {
		return Task{}, err
	}
	if err := os.MkdirAll(r.archiveTasks, 0o755); err != nil {
		return Task{}, fmt.Errorf("board: failed to create archive tasks directory: %w", err)
	}

	path, task, err := r.findTaskFileLockedContext(ctx, r.tasksDir, id)
	if err != nil {
		return Task{}, err
	}

	dest := filepath.Join(r.archiveTasks, filepath.Base(path))
	if err := shared.EnsureInDir(r.archiveTasks, dest); err != nil {
		return Task{}, err
	}
	if err := os.Rename(path, dest); err != nil {
		return Task{}, fmt.Errorf("board: failed to archive task %s: %w", id, err)
	}
	task.FilePath = dest
	r.attachBoardInfo(&task)
	return task, nil
}

// RestoreTask moves an archived task back to the active tasks directory.
func (r *Repository) RestoreTask(id string) (Task, error) {
	return r.RestoreTaskContext(context.Background(), id)
}

// RestoreTaskContext moves an archived task back to the active tasks directory, honoring ctx cancellation.
func (r *Repository) RestoreTaskContext(ctx context.Context, id string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	select {
	case <-ctx.Done():
		return Task{}, ctx.Err()
	default:
	}
	if err := validateID(id); err != nil {
		return Task{}, err
	}
	if err := ensureDirExists(r.archiveTasks); err != nil {
		return Task{}, err
	}
	if err := os.MkdirAll(r.tasksDir, 0o755); err != nil {
		return Task{}, fmt.Errorf("board: failed to create tasks directory: %w", err)
	}

	path, task, err := r.findTaskFileLockedContext(ctx, r.archiveTasks, id)
	if err != nil {
		return Task{}, err
	}

	dest := filepath.Join(r.tasksDir, filepath.Base(path))
	if err := shared.EnsureInDir(r.tasksDir, dest); err != nil {
		return Task{}, err
	}
	if err := os.Rename(path, dest); err != nil {
		return Task{}, fmt.Errorf("board: failed to restore task %s: %w", id, err)
	}
	task.FilePath = dest
	r.attachBoardInfo(&task)
	return task, nil
}

// ListArchivedTasks returns all archived tasks.
func (r *Repository) ListArchivedTasks() ([]Task, error) {
	return r.ListArchivedTasksContext(context.Background())
}

// ListArchivedTasksContext returns all archived tasks, honoring ctx cancellation.
func (r *Repository) ListArchivedTasksContext(ctx context.Context) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := ensureDirExists(r.archiveTasks); err != nil {
		return nil, err
	}
	return r.readTasksFromDirContext(ctx, r.archiveTasks)
}

// ArchiveBefore moves tasks created before the given date.
func (r *Repository) ArchiveBefore(cutoff time.Time) ([]Task, error) {
	return r.ArchiveBeforeContext(context.Background(), cutoff)
}

// ArchiveBeforeContext moves tasks created before the given date, honoring ctx cancellation.
func (r *Repository) ArchiveBeforeContext(ctx context.Context, cutoff time.Time) ([]Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ensureDirExists(r.tasksDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(r.archiveTasks, 0o755); err != nil {
		return nil, fmt.Errorf("board: failed to create archive tasks directory: %w", err)
	}

	tasks, err := r.readTasksFromDirContext(ctx, r.tasksDir)
	if err != nil {
		return nil, err
	}

	var moved []Task
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if task.Created.IsZero() || !task.Created.Before(cutoff) {
			continue
		}
		src := filepath.Join(r.tasksDir, filepath.Base(task.FilePath))
		dest := filepath.Join(r.archiveTasks, filepath.Base(task.FilePath))
		if err := shared.EnsureInDir(r.tasksDir, src); err != nil {
			return nil, err
		}
		if err := shared.EnsureInDir(r.archiveTasks, dest); err != nil {
			return nil, err
		}
		if err := os.Rename(src, dest); err != nil {
			return nil, fmt.Errorf("board: failed to archive task %s: %w", task.ID, err)
		}
		task.FilePath = dest
		moved = append(moved, task)
	}
	return moved, nil
}

// DeleteArchivedTask removes an archived task permanently.
func (r *Repository) DeleteArchivedTask(id string) error {
	return r.DeleteArchivedTaskContext(context.Background(), id)
}

// DeleteArchivedTaskContext removes an archived task permanently, honoring ctx cancellation.
func (r *Repository) DeleteArchivedTaskContext(ctx context.Context, id string) error {
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
	if err := ensureDirExists(r.archiveTasks); err != nil {
		return err
	}
	path, _, err := r.findTaskFileLockedContext(ctx, r.archiveTasks, id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("board: failed to delete archived task %s: %w", id, err)
	}
	return nil
}

// DeleteTask removes an active task permanently.
func (r *Repository) DeleteTask(id string) error {
	return r.DeleteTaskContext(context.Background(), id)
}

// DeleteTaskContext removes an active task permanently, honoring ctx cancellation.
func (r *Repository) DeleteTaskContext(ctx context.Context, id string) error {
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
	path, _, err := r.findTaskFileLockedContext(ctx, r.tasksDir, id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("board: failed to delete task %s: %w", id, err)
	}
	return nil
}

func (r *Repository) findTaskFileLockedContext(ctx context.Context, dir, id string) (string, Task, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", Task{}, fmt.Errorf("board: failed to read tasks directory: %w", err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return "", Task{}, ctx.Err()
		default:
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := shared.EnsureInDir(dir, path); err != nil {
			return "", Task{}, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", Task{}, fmt.Errorf("board: failed to read task file %s: %w", path, err)
		}
		task, err := r.parser.Parse(data)
		if err != nil {
			return "", Task{}, fmt.Errorf("board: failed to parse task file %s: %w", path, err)
		}
		if task.ID != id {
			continue
		}
		task.FilePath = path
		return path, task, nil
	}
	return "", Task{}, fmt.Errorf("board: %w", ErrTaskNotFound)
}

func (r *Repository) readTasksFromDirContext(ctx context.Context, dir string) ([]Task, error) {
	entries, err := os.ReadDir(dir)
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
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := shared.EnsureInDir(dir, path); err != nil {
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
