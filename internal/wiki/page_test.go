package wiki

import (
	"errors"
	"testing"
)

func TestParsePage(t *testing.T) {
	valid := `---
title: "Architecture Overview"
slug: "architecture/overview"
section: "Architecture"
order: 10
tags: [architecture, core]
status: "published"
---
# Title
Body line
`

	cases := []struct {
		name    string
		input   string
		want    Page
		wantErr error
	}{
		{
			name:  "valid frontmatter",
			input: valid,
			want: Page{
				Title:   "Architecture Overview",
				Slug:    "architecture/overview",
				Section: "Architecture",
				Order:   10,
				Tags:    []string{"architecture", "core"},
				Status:  "published",
				Content: "# Title\nBody line",
			},
		},
		{
			name:    "missing frontmatter start",
			input:   "title: test\n---\nbody\n",
			wantErr: ErrInvalidFrontmatter,
		},
		{
			name:    "missing frontmatter end",
			input:   "---\ntitle: test\nbody\n",
			wantErr: ErrInvalidFrontmatter,
		},
		{
			name:    "invalid yaml",
			input:   "---\nslug: [\n---\nbody\n",
			wantErr: ErrInvalidYAML,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := []byte(tc.input)

			// Act
			got, err := ParsePage(input)

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

			if got.Title != tc.want.Title || got.Slug != tc.want.Slug || got.Section != tc.want.Section || got.Order != tc.want.Order {
				t.Fatalf("unexpected basic fields: %+v", got)
			}
			if got.Status != tc.want.Status {
				t.Fatalf("unexpected status: %s", got.Status)
			}
			if got.Content != tc.want.Content {
				t.Fatalf("unexpected content: %q", got.Content)
			}
			if len(got.Tags) != len(tc.want.Tags) {
				t.Fatalf("unexpected tags: %+v", got.Tags)
			}
			for i := range got.Tags {
				if got.Tags[i] != tc.want.Tags[i] {
					t.Fatalf("unexpected tags: %+v", got.Tags)
				}
			}
		})
	}
}
