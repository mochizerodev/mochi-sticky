package wiki

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ExportSelection scopes wiki exports to a page or section.
type ExportSelection struct {
	Page    string
	Section string
}

// DefaultExportPath builds a default export path based on selection and format.
func DefaultExportPath(root, format string, selection ExportSelection) string {
	ext := "md"
	if strings.EqualFold(strings.TrimSpace(format), "pdf") {
		ext = "pdf"
	}
	page := strings.TrimSpace(selection.Page)
	section := strings.TrimSpace(selection.Section)
	switch {
	case page != "":
		return filepath.Join(root, fmt.Sprintf("export-page-%s.%s", exportName(page), ext))
	case section != "":
		return filepath.Join(root, fmt.Sprintf("export-section-%s.%s", exportName(section), ext))
	default:
		return filepath.Join(root, fmt.Sprintf("export.%s", ext))
	}
}

func exportName(value string) string {
	name := Slugify(strings.TrimSpace(value))
	if name == "" {
		return "selection"
	}
	return name
}

// BuildPageManifest returns a manifest containing a single page.
func BuildPageManifest(root string, pages []Page, slug string) ([]ManifestEntry, error) {
	normalized, err := NormalizeSlug(slug)
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		pageSlug := strings.TrimSpace(page.Slug)
		if pageSlug == "" && strings.TrimSpace(page.FilePath) != "" {
			pageSlug = SlugFromPath(root, page.FilePath)
		}
		if pageSlug != normalized {
			continue
		}
		if shouldSkipStatus(page.Status) {
			return nil, fmt.Errorf("wiki: page %s is %s", normalized, strings.TrimSpace(page.Status))
		}
		title := strings.TrimSpace(page.Title)
		if title == "" {
			title = normalized
		}
		return []ManifestEntry{{
			Title: title,
			Slug:  normalized,
			Order: page.Order,
		}}, nil
	}
	return nil, ErrPageNotFound
}

// BuildSectionManifest returns a manifest containing pages in a specific section.
func BuildSectionManifest(index Index, pages []Page, section string) ([]ManifestEntry, error) {
	target, _, err := FindSection(index, section)
	if err != nil {
		return nil, err
	}
	pageBySlug := make(map[string]Page, len(pages))
	for _, page := range pages {
		slug := strings.TrimSpace(page.Slug)
		if slug == "" {
			continue
		}
		pageBySlug[slug] = page
	}
	manifest := make([]ManifestEntry, 0)
	for _, fullSlug := range target.ListPages() {
		page, ok := pageBySlug[fullSlug]
		if !ok {
			return nil, ErrPageNotFound
		}
		if shouldSkipStatus(page.Status) {
			continue
		}
		title := strings.TrimSpace(page.Title)
		if title == "" {
			title = fullSlug
		}
		manifest = append(manifest, ManifestEntry{
			Title: title,
			Slug:  fullSlug,
			Order: page.Order,
		})
	}
	return manifest, nil
}
