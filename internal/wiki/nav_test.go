package wiki

import "testing"

func TestBuildNavTree(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{
				Title: "Architecture",
				Slug:  "architecture",
				Order: 1,
				Pages: []string{"overview", "decisions"},
			},
		},
	}

	pages := []Page{
		{Title: "Overview", Slug: "architecture/overview", Order: 1},
		{Title: "Decisions", Slug: "architecture/decisions", Order: 2},
	}

	// Act
	tree, err := BuildNavTree(index, pages)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tree) != 1 {
		t.Fatalf("expected 1 section, got %d", len(tree))
	}
	if len(tree[0].Pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(tree[0].Pages))
	}
	if tree[0].Pages[0].Slug != "architecture/overview" {
		t.Fatalf("unexpected slug: %s", tree[0].Pages[0].Slug)
	}
	if tree[0].Slug != "architecture" || tree[0].Title != "Architecture" {
		t.Fatalf("unexpected section: %+v", tree[0])
	}
}

func TestBuildNavTreeMissingPage(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{
				Title: "Architecture",
				Slug:  "architecture",
				Pages: []string{"overview"},
			},
		},
	}

	// Act
	_, err := BuildNavTree(index, []Page{})

	// Assert
	if err == nil {
		t.Fatalf("expected error for missing page")
	}
}
