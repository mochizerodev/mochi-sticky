package wiki

import (
	"sort"
	"strings"
)

// ManifestEntry represents a flattened page entry for export.
type ManifestEntry struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
	Order int    `json:"order"`
}

// BuildManifest flattens the index into an ordered list of pages.
func BuildManifest(index Index, pages []Page) ([]ManifestEntry, error) {
	pageBySlug := make(map[string]Page, len(pages))
	for _, page := range pages {
		if strings.TrimSpace(page.Slug) == "" {
			continue
		}
		pageBySlug[page.Slug] = page
	}

	manifest := make([]ManifestEntry, 0)
	for _, section := range index.Sections {
		for _, fullSlug := range section.ListPages() {
			page, ok := pageBySlug[fullSlug]
			if !ok {
				return nil, ErrPageNotFound
			}
			if shouldSkipStatus(page.Status) {
				continue
			}
			manifest = append(manifest, ManifestEntry{
				Title: page.Title,
				Slug:  fullSlug,
				Order: page.Order,
			})
		}
	}
	return manifest, nil
}

// BuildManifestFromPages builds a manifest without an index, sorted by title.
func BuildManifestFromPages(pages []Page) []ManifestEntry {
	manifest := make([]ManifestEntry, 0)
	for _, page := range pages {
		if strings.TrimSpace(page.Slug) == "" {
			continue
		}
		if shouldSkipStatus(page.Status) {
			continue
		}
		manifest = append(manifest, ManifestEntry{
			Title: page.Title,
			Slug:  page.Slug,
			Order: page.Order,
		})
	}
	sort.Slice(manifest, func(i, j int) bool {
		return strings.ToLower(manifest[i].Title) < strings.ToLower(manifest[j].Title)
	})
	return manifest
}
