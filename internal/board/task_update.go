package board

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// UpdateTaskTitle updates a task title by ID.
func (r *Repository) UpdateTaskTitle(id, title string) error {
	return r.UpdateTaskTitleContext(context.Background(), id, title)
}

// UpdateTaskTitleContext updates a task title by ID, honoring ctx cancellation.
func (r *Repository) UpdateTaskTitleContext(ctx context.Context, id, title string) error {
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
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return fmt.Errorf("board: %w", ErrInvalidTitle)
	}
	return r.updateTaskLockedContext(ctx, id, func(task *Task) error {
		task.Title = trimmed
		return nil
	})
}

// UpdateTaskTags updates a task's tags by ID.
func (r *Repository) UpdateTaskTags(id string, tags []string) error {
	return r.UpdateTaskTagsContext(context.Background(), id, tags)
}

// UpdateTaskTagsContext updates a task's tags by ID, honoring ctx cancellation.
func (r *Repository) UpdateTaskTagsContext(ctx context.Context, id string, tags []string) error {
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
	normalized := NormalizeTags(tags)
	return r.updateTaskLockedContext(ctx, id, func(task *Task) error {
		task.Tags = normalized
		return nil
	})
}

// UpdateTaskPriority updates a task's priority by ID.
func (r *Repository) UpdateTaskPriority(id string, priority int) error {
	return r.UpdateTaskPriorityContext(context.Background(), id, priority)
}

// UpdateTaskPriorityContext updates a task's priority by ID, honoring ctx cancellation.
func (r *Repository) UpdateTaskPriorityContext(ctx context.Context, id string, priority int) error {
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
	normalized, err := normalizePriority(priority)
	if err != nil {
		return err
	}
	return r.updateTaskLockedContext(ctx, id, func(task *Task) error {
		task.Priority = normalized
		return nil
	})
}

// UpdateTaskContent updates a task's markdown content by ID.
func (r *Repository) UpdateTaskContent(id, content string) error {
	return r.UpdateTaskContentContext(context.Background(), id, content)
}

// UpdateTaskContentContext updates a task's markdown content by ID, honoring ctx cancellation.
func (r *Repository) UpdateTaskContentContext(ctx context.Context, id, content string) error {
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
	return r.updateTaskLockedContext(ctx, id, func(task *Task) error {
		task.Content = content
		return nil
	})
}

func (r *Repository) updateTaskLockedContext(ctx context.Context, id string, update func(*Task) error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := ensureDirExists(r.tasksDir); err != nil {
		return err
	}
	path, task, err := r.findTaskFileLockedContext(ctx, r.tasksDir, id)
	if err != nil {
		return err
	}
	if err := update(&task); err != nil {
		return err
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
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("board: failed to write task file %s: %w", path, err)
	}
	return nil
}
