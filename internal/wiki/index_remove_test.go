package wiki

import "testing"

func TestIndexRemoveSlug(t *testing.T) {
	// Arrange
	index := Index{
		Sections: []IndexSection{
			{Slug: "architecture", Pages: []string{"overview", "decisions"}},
			{Slug: "", Pages: []string{"root-page"}},
		},
	}

	// Act
	removedArchitecture := index.RemoveSlug("architecture/overview")
	archPages := append([]string(nil), index.Sections[0].Pages...)
	removedRoot := index.RemoveSlug("root-page")
	rootPages := append([]string(nil), index.Sections[1].Pages...)

	// Assert
	if !removedArchitecture {
		t.Fatalf("expected removal")
	}
	if len(archPages) != 1 || archPages[0] != "decisions" {
		t.Fatalf("unexpected pages: %+v", archPages)
	}
	if !removedRoot {
		t.Fatalf("expected root removal")
	}
	if len(rootPages) != 0 {
		t.Fatalf("unexpected root pages: %+v", rootPages)
	}
}
