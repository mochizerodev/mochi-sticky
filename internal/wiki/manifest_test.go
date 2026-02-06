package wiki

import "testing"

func TestBuildManifest(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "Ops", Slug: "ops", Pages: []string{"runbooks", "release"}},
		},
	}
	pages := []Page{
		{Title: "Runbooks", Slug: "ops/runbooks", Order: 1, Status: "published"},
		{Title: "Release", Slug: "ops/release", Order: 2, Status: "draft"},
	}

	// Act
	manifest, err := BuildManifest(index, pages)

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

func TestBuildManifestFromPages(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "B", Slug: "b", Status: "published"},
		{Title: "A", Slug: "a", Status: "published"},
		{Title: "C", Slug: "c", Status: "archived"},
	}

	// Act
	manifest := BuildManifestFromPages(pages)

	// Assert
	if len(manifest) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(manifest))
	}
	if manifest[0].Slug != "a" || manifest[1].Slug != "b" {
		t.Fatalf("unexpected order: %+v", manifest)
	}
}
