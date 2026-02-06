package wiki

import "testing"

func TestFindSectionBySlug(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "Home", Slug: ""},
			{Title: "Ops", Slug: "ops"},
		},
	}

	// Act
	section, _, err := FindSection(index, "ops")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section.Slug != "ops" {
		t.Fatalf("unexpected section: %+v", section)
	}
}

func TestFindSectionByTitle(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "Architecture", Slug: "architecture"},
		},
	}

	// Act
	section, _, err := FindSection(index, "Architecture")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section.Slug != "architecture" {
		t.Fatalf("unexpected section: %+v", section)
	}
}

func TestResolveLinkedSections(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "A", Slug: "a"},
			{Title: "B", Slug: "b"},
		},
	}
	section := IndexSection{
		Title: "A",
		Slug:  "a",
		Links: SectionLinks{DependsOn: []string{"b"}},
	}

	// Act
	linked, err := ResolveLinkedSections(index, section, []string{"depends_on"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(linked) != 1 || linked[0].Slug != "b" {
		t.Fatalf("unexpected linked sections: %+v", linked)
	}
}

func TestBuildManifestForSections(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Title: "Ops", Slug: "ops", Pages: []string{"runbooks"}},
			{Title: "Dev", Slug: "dev", Pages: []string{"guide"}},
		},
	}
	pages := []Page{
		{Title: "Runbooks", Slug: "ops/runbooks", Status: "published"},
		{Title: "Guide", Slug: "dev/guide", Status: "published"},
	}

	// Act
	manifest, err := BuildManifestForSections(index, pages, []IndexSection{index.Sections[1], index.Sections[0]})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest) != 2 || manifest[0].Slug != "dev/guide" || manifest[1].Slug != "ops/runbooks" {
		t.Fatalf("unexpected manifest order: %+v", manifest)
	}
}

func TestFilterSectionsByTagAndLink(t *testing.T) {
	// Arrange
	sections := []IndexSection{
		{
			Title: "Ops",
			Slug:  "ops",
			Tags:  []string{"infra"},
			Links: SectionLinks{RelatedTo: []string{"dev"}},
		},
		{Title: "Dev", Slug: "dev", Tags: []string{"code"}},
	}

	// Act
	filtered := FilterSections(sections, SectionFilterOptions{
		Tags:       []string{"infra"},
		TagMode:    "any",
		LinkType:   "related_to",
		LinkTarget: "dev",
	})

	// Assert
	if len(filtered) != 1 || filtered[0].Slug != "ops" {
		t.Fatalf("unexpected filter result: %+v", filtered)
	}
}
