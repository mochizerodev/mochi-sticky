package board

import (
	"sort"
	"strings"
	"time"
)

// ListOptions controls filtering and sorting of tasks.
// ListOptions controls filtering and sorting of tasks.
type ListOptions struct {
	Status  string
	Title   string
	Tags    []string
	TagMode string
	From    time.Time
	To      time.Time
	SortBy  string
	Desc    bool
}

// FilterAndSortTasks applies list options to tasks.
// FilterAndSortTasks applies the provided ListOptions and then sorts the results.
func FilterAndSortTasks(tasks []Task, opts ListOptions) []Task {
	filtered := filterTasks(tasks, opts)
	sortTasks(filtered, opts)
	return filtered
}

func filterTasks(tasks []Task, opts ListOptions) []Task {
	status := strings.TrimSpace(opts.Status)
	title := strings.TrimSpace(opts.Title)
	if status == "" && title == "" && len(opts.Tags) == 0 && opts.From.IsZero() && opts.To.IsZero() {
		return append([]Task(nil), tasks...)
	}

	statusLower := strings.ToLower(status)
	titleLower := strings.ToLower(title)
	tagMode := strings.ToLower(strings.TrimSpace(opts.TagMode))
	tagFilters := normalizeTagFilters(opts.Tags)

	filtered := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if statusLower != "" && strings.ToLower(task.Status) != statusLower {
			continue
		}
		if titleLower != "" && !strings.Contains(strings.ToLower(task.Title), titleLower) {
			continue
		}
		if len(tagFilters) > 0 && !matchesTags(task.Tags, tagFilters, tagMode) {
			continue
		}
		if !matchesDateRange(task.Created, opts.From, opts.To) {
			continue
		}
		filtered = append(filtered, task)
	}
	return filtered
}

func sortTasks(tasks []Task, opts ListOptions) {
	sortKey := strings.ToLower(strings.TrimSpace(opts.SortBy))
	if sortKey == "" {
		return
	}

	less := func(i, j int) bool {
		switch sortKey {
		case "status":
			return strings.ToLower(tasks[i].Status) < strings.ToLower(tasks[j].Status)
		case "created":
			return tasks[i].Created.Before(tasks[j].Created.Time)
		case "priority":
			left := effectivePriority(tasks[i].Priority)
			right := effectivePriority(tasks[j].Priority)
			if left == right {
				return strings.ToLower(tasks[i].Title) < strings.ToLower(tasks[j].Title)
			}
			return left < right
		case "title":
			return strings.ToLower(tasks[i].Title) < strings.ToLower(tasks[j].Title)
		default:
			return tasks[i].ID < tasks[j].ID
		}
	}

	if opts.Desc {
		sort.SliceStable(tasks, func(i, j int) bool { return !less(i, j) })
		return
	}
	sort.SliceStable(tasks, less)
}

func normalizeTagFilters(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, strings.ToLower(trimmed))
	}
	return normalized
}

func matchesTags(taskTags []string, filters []string, mode string) bool {
	if len(filters) == 0 {
		return true
	}
	tagSet := make(map[string]struct{}, len(taskTags))
	for _, tag := range taskTags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		tagSet[strings.ToLower(trimmed)] = struct{}{}
	}
	if len(tagSet) == 0 {
		return false
	}

	if mode == "all" {
		for _, filter := range filters {
			if _, ok := tagSet[filter]; !ok {
				return false
			}
		}
		return true
	}

	for _, filter := range filters {
		if _, ok := tagSet[filter]; ok {
			return true
		}
	}
	return false
}

func matchesDateRange(value Date, from, to time.Time) bool {
	if value.IsZero() {
		return from.IsZero() && to.IsZero()
	}
	if !from.IsZero() && value.Before(from) {
		return false
	}
	if !to.IsZero() && value.After(to) {
		return false
	}
	return true
}
