package wiki

import (
	"errors"
	"testing"
)

func TestRelatedPages(t *testing.T) {
	// Arrange
	pages := []Page{
		{Title: "One", Slug: "one", Tags: []string{"Go", "Wiki"}},
		{Title: "Two", Slug: "two", Tags: []string{"go", "cli"}},
		{Title: "Three", Slug: "three", Tags: []string{"wiki", "nav"}, Status: "draft"},
		{Title: "Four", Slug: "four", Tags: []string{"wiki", "go"}},
	}

	// Act
	results, err := RelatedPages(pages, "one")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Slug != "four" || results[0].SharedCount != 2 {
		t.Fatalf("unexpected top result: %+v", results[0])
	}
	if results[1].Slug != "two" || results[1].SharedCount != 1 {
		t.Fatalf("unexpected second result: %+v", results[1])
	}
}

func TestRelatedPagesMissingSlug(t *testing.T) {
	// Arrange
	pages := []Page{}

	// Act
	_, err := RelatedPages(pages, "")

	// Assert
	if !errors.Is(err, ErrPageNotFound) {
		t.Fatalf("expected ErrPageNotFound, got %v", err)
	}
}
