package wiki

import "errors"

import "testing"

func TestValidateUniqueSlugs(t *testing.T) {
	cases := []struct {
		name    string
		pages   []Page
		wantErr error
	}{
		{
			name: "unique slugs",
			pages: []Page{
				{Slug: "a"},
				{Slug: "b"},
			},
		},
		{
			name: "duplicate slugs",
			pages: []Page{
				{Slug: "a"},
				{Slug: "a"},
			},
			wantErr: ErrDuplicateSlug,
		},
		{
			name: "empty slug ignored",
			pages: []Page{
				{Slug: ""},
				{Slug: "a"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			pages := tc.pages

			// Act
			err := ValidateUniqueSlugs(pages)

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
		})
	}
}
