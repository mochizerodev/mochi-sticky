package wiki

import "testing"

func TestLintPages(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "", Slug: "empty", Content: "No heading"},
		{Title: "Has Heading", Slug: "ok", Content: "# Heading\nBody"},
		{Title: "Bad Link", Slug: "bad-link", Content: "# Title\nSee [link]( )"},
	}

	// Act
	issues := LintPages(pages)

	// Assert
	if len(issues) == 0 {
		t.Fatalf("expected issues")
	}
	foundBadLink := false
	for _, issue := range issues {
		if issue.Slug == "bad-link" && issue.Message != "" {
			foundBadLink = true
		}
	}
	if !foundBadLink {
		t.Fatalf("expected bad link issue")
	}
}
