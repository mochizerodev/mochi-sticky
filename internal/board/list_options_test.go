package board

import (
	"testing"
	"time"
)

func TestFilterAndSortTasks(t *testing.T) {
	// Arrange
	baseDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tasks := []Task{
		{ID: "T-1", Title: "Alpha", Status: "todo", Priority: 2, Tags: []string{"backend"}, Created: Date{Time: baseDate}},
		{ID: "T-2", Title: "Bravo", Status: "doing", Priority: 1, Tags: []string{"frontend", "ui"}, Created: Date{Time: baseDate.AddDate(0, 0, 1)}},
		{ID: "T-3", Title: "Charlie", Status: "done", Priority: 3, Tags: []string{"backend", "api"}, Created: Date{Time: baseDate.AddDate(0, 0, 2)}},
		{ID: "T-4", Title: "Delta", Status: "todo", Priority: 2, Tags: nil, Created: Date{Time: baseDate.AddDate(0, 0, 3)}},
	}

	cases := []struct {
		name   string
		opts   ListOptions
		expect []string
	}{
		{
			name:   "status filter",
			opts:   ListOptions{Status: "todo"},
			expect: []string{"T-1", "T-4"},
		},
		{
			name:   "title filter",
			opts:   ListOptions{Title: "brav"},
			expect: []string{"T-2"},
		},
		{
			name:   "tags any",
			opts:   ListOptions{Tags: []string{"backend"}},
			expect: []string{"T-1", "T-3"},
		},
		{
			name:   "tags all",
			opts:   ListOptions{Tags: []string{"backend", "api"}, TagMode: "all"},
			expect: []string{"T-3"},
		},
		{
			name:   "date range from",
			opts:   ListOptions{From: baseDate.AddDate(0, 0, 2)},
			expect: []string{"T-3", "T-4"},
		},
		{
			name:   "date range to",
			opts:   ListOptions{To: baseDate.AddDate(0, 0, 1)},
			expect: []string{"T-1", "T-2"},
		},
		{
			name:   "date range between",
			opts:   ListOptions{From: baseDate.AddDate(0, 0, 1), To: baseDate.AddDate(0, 0, 2)},
			expect: []string{"T-2", "T-3"},
		},
		{
			name:   "combined filters",
			opts:   ListOptions{Status: "todo", Tags: []string{"backend"}, From: baseDate},
			expect: []string{"T-1"},
		},
		{
			name:   "sort by priority",
			opts:   ListOptions{SortBy: "priority"},
			expect: []string{"T-2", "T-1", "T-4", "T-3"},
		},
		{
			name:   "sort by priority desc",
			opts:   ListOptions{SortBy: "priority", Desc: true},
			expect: []string{"T-3", "T-4", "T-1", "T-2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			opts := tc.opts
			expect := tc.expect

			// Act
			got := FilterAndSortTasks(tasks, opts)

			// Assert
			if len(got) != len(expect) {
				t.Fatalf("expected %d tasks, got %d", len(expect), len(got))
			}
			for i, task := range got {
				if task.ID != expect[i] {
					t.Fatalf("expected %s at %d, got %s", expect[i], i, task.ID)
				}
			}
		})
	}
}
