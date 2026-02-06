package board

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestParserParse(t *testing.T) {
	parser := &Parser{}

	valid := `---
id: "task-1"
title: "Test Task"
status: "todo"
priority: 2
tags: [one, two]
created: 2026-01-29
---
# Title
Body line
`

	cases := []struct {
		name    string
		input   string
		want    Task
		wantErr error
	}{
		{
			name:  "valid frontmatter",
			input: valid,
			want: Task{
				ID:       "task-1",
				Title:    "Test Task",
				Status:   "todo",
				Priority: 2,
				Tags:     []string{"one", "two"},
				Created:  Date{Time: time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)},
				Content:  "# Title\nBody line",
			},
		},
		{
			name:    "missing frontmatter start",
			input:   "id: 1\n---\nbody\n",
			wantErr: ErrInvalidFrontmatter,
		},
		{
			name:    "missing frontmatter end",
			input:   "---\nid: 1\nbody\n",
			wantErr: ErrInvalidFrontmatter,
		},
		{
			name:    "invalid yaml",
			input:   "---\nid: [\n---\nbody\n",
			wantErr: ErrInvalidYAML,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := []byte(tc.input)

			// Act
			got, err := parser.Parse(input)

			// Assert
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.ID != tc.want.ID || got.Title != tc.want.Title || got.Status != tc.want.Status || got.Priority != tc.want.Priority {
				t.Fatalf("unexpected basic fields: %+v", got)
			}
			if !reflect.DeepEqual(got.Tags, tc.want.Tags) {
				t.Fatalf("unexpected tags: %+v", got.Tags)
			}
			if !got.Created.Equal(tc.want.Created.Time) {
				t.Fatalf("unexpected created date: %v", got.Created)
			}
			if got.Content != tc.want.Content {
				t.Fatalf("unexpected content: %q", got.Content)
			}
		})
	}
}
