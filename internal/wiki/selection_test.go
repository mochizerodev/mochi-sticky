package wiki

import "testing"

func TestBuildPageManifest(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "One", Slug: "section/one", Status: "published"},
	}

	// Act
	manifest, err := BuildPageManifest("/tmp", pages, "section/one")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(manifest))
	}
	if manifest[0].Slug != "section/one" {
		t.Fatalf("unexpected slug: %s", manifest[0].Slug)
	}
}

func TestBuildPageManifestDraft(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "Draft", Slug: "draft", Status: "draft"},
	}

	// Act
	_, err := BuildPageManifest("/tmp", pages, "draft")

	// Assert
	if err == nil {
		t.Fatalf("expected error for draft page")
	}
}

func TestBuildSectionManifestByTitle(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "Development", Slug: "development", Pages: []string{"a", "b"}},
		},
	}
	pages := []Page{
		{Title: "A", Slug: "development/a", Status: "published"},
		{Title: "B", Slug: "development/b", Status: "published"},
	}

	// Act
	manifest, err := BuildSectionManifest(index, pages, "Development")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(manifest))
	}
	if manifest[0].Slug != "development/a" || manifest[1].Slug != "development/b" {
		t.Fatalf("unexpected order: %+v", manifest)
	}
}

func TestBuildSectionManifestEmptySlug(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "Home", Slug: "", Pages: []string{"home"}},
		},
	}
	pages := []Page{
		{Title: "Home", Slug: "home", Status: "published"},
	}

	// Act
	manifest, err := BuildSectionManifest(index, pages, "Home")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest) != 1 || manifest[0].Slug != "home" {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
}

func TestBuildSectionManifestSkipsDraft(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "Ops", Slug: "ops", Pages: []string{"runbooks", "draft"}},
		},
	}
	pages := []Page{
		{Title: "Runbooks", Slug: "ops/runbooks", Status: "published"},
		{Title: "Draft", Slug: "ops/draft", Status: "draft"},
	}

	// Act
	manifest, err := BuildSectionManifest(index, pages, "ops")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(manifest))
	}
	if manifest[0].Slug != "ops/runbooks" {
		t.Fatalf("unexpected slug: %s", manifest[0].Slug)
	}
}
