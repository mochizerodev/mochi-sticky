package board

import (
	"fmt"
)

// IsReady reports whether all dependencies are satisfied (dependents exist and are done).
// It also returns the list of unmet dependency IDs.
func IsReady(task Task, index map[string]Task) (bool, []string) {
	if len(task.DependsOn) == 0 {
		return true, nil
	}
	var unmet []string
	for _, dep := range task.DependsOn {
		depTask, ok := index[dep]
		if !ok || !isDoneStatus(depTask.Status) {
			unmet = append(unmet, dep)
		}
	}
	return len(unmet) == 0, unmet
}

// ValidateNoCycles ensures the dependency graph has no cycles.
func ValidateNoCycles(tasks []Task) error {
	index := make(map[string][]string, len(tasks))
	for _, t := range tasks {
		index[t.ID] = normalizeIDs(t.DependsOn)
	}
	visited := make(map[string]int) // 0=unseen,1=visiting,2=done
	var dfs func(string) error
	dfs = func(id string) error {
		state := visited[id]
		if state == 1 {
			return fmt.Errorf("board: dependency cycle detected at %s: %w", id, ErrInvalidDependency)
		}
		if state == 2 {
			return nil
		}
		visited[id] = 1
		for _, dep := range index[id] {
			if err := dfs(dep); err != nil {
				return err
			}
		}
		visited[id] = 2
		return nil
	}
	for id := range index {
		if err := dfs(id); err != nil {
			return err
		}
	}
	return nil
}

func isDoneStatus(status string) bool {
	switch normalized := normalizeStatus(status); normalized {
	case "done", "archived":
		return true
	default:
		return false
	}
}

func normalizeStatus(status string) string {
	return slugify(status)
}
