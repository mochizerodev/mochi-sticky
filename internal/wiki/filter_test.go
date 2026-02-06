package wiki

import "testing"

func TestFilterPagesByTitle(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "Architecture Overview", Slug: "architecture/overview"},
		{Title: "Runbooks", Slug: "ops/runbooks"},
	}

	// Act
	filtered := FilterPages(pages, FilterOptions{Title: "arch", CaseInsensitive: true})

	// Assert
	if len(filtered) != 1 || filtered[0].Slug != "architecture/overview" {
		t.Fatalf("unexpected filter result: %+v", filtered)
	}
}

func TestFilterPagesByTagsAny(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "A", Slug: "a", Tags: []string{"wiki", "ops"}},
		{Title: "B", Slug: "b", Tags: []string{"dev"}},
	}

	// Act
	filtered := FilterPages(pages, FilterOptions{Tags: []string{"OPS"}, TagMode: "any", CaseInsensitive: true})

	// Assert
	if len(filtered) != 1 || filtered[0].Slug != "a" {
		t.Fatalf("unexpected tag filter result: %+v", filtered)
	}
}

func TestFilterPagesByTagsAll(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "A", Slug: "a", Tags: []string{"wiki", "ops"}},
		{Title: "B", Slug: "b", Tags: []string{"wiki"}},
	}

	// Act
	filtered := FilterPages(pages, FilterOptions{Tags: []string{"wiki", "ops"}, TagMode: "all", CaseInsensitive: true})

	// Assert
	if len(filtered) != 1 || filtered[0].Slug != "a" {
		t.Fatalf("unexpected tag filter result: %+v", filtered)
	}
}

func TestFilterPagesBySection(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "Guide", Slug: "user-guide/wiki", Section: "User Guide"},
		{Title: "Ref", Slug: "reference/config", Section: "Reference"},
	}

	// Act
	filtered := FilterPages(pages, FilterOptions{Section: "user-guide", CaseInsensitive: true})

	// Assert
	if len(filtered) != 1 || filtered[0].Slug != "user-guide/wiki" {
		t.Fatalf("unexpected section filter result: %+v", filtered)
	}
}

func TestFilterPagesByQuery(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "Caching", Slug: "ops/cache", Tags: []string{"performance"}, Content: "Redis runbook"},
		{Title: "Auth", Slug: "security/auth", Tags: []string{"security"}},
	}

	// Act
	filtered := FilterPages(pages, FilterOptions{Query: "redis runbook", CaseInsensitive: true})

	// Assert
	if len(filtered) != 1 || filtered[0].Slug != "ops/cache" {
		t.Fatalf("unexpected query filter result: %+v", filtered)
	}
}

func TestFilterManifest(t *testing.T) {
	// Arrange
	manifest := []ManifestEntry{
		{Title: "A", Slug: "a"},
		{Title: "B", Slug: "b"},
	}
	pages := map[string]Page{
		"a": {Title: "A", Slug: "a", Tags: []string{"keep"}},
		"b": {Title: "B", Slug: "b", Tags: []string{"drop"}},
	}

	// Act
	filtered := FilterManifest(manifest, pages, FilterOptions{Tags: []string{"keep"}, CaseInsensitive: true})

	// Assert
	if len(filtered) != 1 || filtered[0].Slug != "a" {
		t.Fatalf("unexpected manifest filter: %+v", filtered)
	}
}
