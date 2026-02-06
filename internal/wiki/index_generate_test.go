package wiki

import "testing"

func TestGenerateIndex(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "Home", Slug: "home", Section: "Home"},
		{Title: "Overview", Slug: "architecture/overview", Section: "Architecture"},
		{Title: "Decisions", Slug: "architecture/decisions", Section: "Architecture"},
	}

	// Act
	index, err := GenerateIndex(pages)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(index.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(index.Sections))
	}
	if index.Sections[0].Slug != "" {
		t.Fatalf("expected root section first, got %s", index.Sections[0].Slug)
	}
	if len(index.Sections[1].Pages) != 2 {
		t.Fatalf("expected 2 pages in architecture, got %d", len(index.Sections[1].Pages))
	}
}
