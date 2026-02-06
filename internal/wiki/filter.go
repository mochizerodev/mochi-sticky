package wiki

import (
	"strings"
)

// FilterOptions describes wiki page filters.
type FilterOptions struct {
	Title           string
	Tags            []string
	TagMode         string
	Section         string
	Query           string
	CaseInsensitive bool
}

// FilterPages returns pages that match the provided filters.
func FilterPages(pages []Page, opts FilterOptions) []Page {
	if !hasFilters(opts) {
		return pages
	}
	result := make([]Page, 0, len(pages))
	for _, page := range pages {
		if matchesFilters(page, opts) {
			result = append(result, page)
		}
	}
	return result
}

// FilterManifest filters manifest entries using the provided page metadata.
func FilterManifest(manifest []ManifestEntry, pages map[string]Page, opts FilterOptions) []ManifestEntry {
	if !hasFilters(opts) {
		return manifest
	}
	result := make([]ManifestEntry, 0, len(manifest))
	for _, entry := range manifest {
		page, ok := pages[entry.Slug]
		if !ok {
			continue
		}
		if matchesFilters(page, opts) {
			result = append(result, entry)
		}
	}
	return result
}

// SplitQueryTerms breaks a query into lowercase terms.
func SplitQueryTerms(query string) []string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil
	}
	terms := strings.Fields(trimmed)
	for i, term := range terms {
		terms[i] = strings.ToLower(term)
	}
	return terms
}

func hasFilters(opts FilterOptions) bool {
	return strings.TrimSpace(opts.Title) != "" ||
		strings.TrimSpace(opts.Section) != "" ||
		strings.TrimSpace(opts.Query) != "" ||
		len(opts.Tags) > 0
}

func matchesFilters(page Page, opts FilterOptions) bool {
	caseInsensitive := opts.CaseInsensitive
	titleFilter := strings.TrimSpace(opts.Title)
	if titleFilter != "" && !matchText(pageTitle(page), titleFilter, caseInsensitive) {
		return false
	}

	sectionFilter := strings.TrimSpace(opts.Section)
	if sectionFilter != "" && !matchSection(page, sectionFilter, caseInsensitive) {
		return false
	}

	if len(opts.Tags) > 0 && !matchTags(page.Tags, opts.Tags, opts.TagMode) {
		return false
	}

	query := strings.TrimSpace(opts.Query)
	if query != "" && !matchQuery(page, query, caseInsensitive) {
		return false
	}

	return true
}

func pageTitle(page Page) string {
	title := strings.TrimSpace(page.Title)
	if title != "" {
		return title
	}
	return strings.TrimSpace(page.Slug)
}

func matchText(haystack, needle string, caseInsensitive bool) bool {
	if strings.TrimSpace(needle) == "" {
		return true
	}
	if caseInsensitive {
		return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
	}
	return strings.Contains(haystack, needle)
}

func matchSection(page Page, filter string, caseInsensitive bool) bool {
	if strings.TrimSpace(filter) == "" {
		return true
	}
	filterLower := strings.ToLower(strings.TrimSpace(filter))

	sectionTitle := strings.TrimSpace(page.Section)
	if sectionTitle != "" {
		if strings.ToLower(sectionTitle) == filterLower {
			return true
		}
		if caseInsensitive && strings.Contains(strings.ToLower(sectionTitle), filterLower) {
			return true
		}
	}

	sectionSlug := pageSectionSlug(page)
	if sectionSlug != "" && strings.ToLower(sectionSlug) == filterLower {
		return true
	}
	return false
}

func pageSectionSlug(page Page) string {
	slug := strings.TrimSpace(page.Slug)
	if slug == "" {
		return ""
	}
	parts := strings.SplitN(slug, "/", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func matchTags(pageTags, filterTags []string, mode string) bool {
	normalizedPage := normalizeTagSet(pageTags)
	normalizedFilter := normalizeTagSet(filterTags)
	if len(normalizedFilter) == 0 {
		return true
	}
	requireAll := strings.EqualFold(strings.TrimSpace(mode), "all")
	if requireAll {
		for tag := range normalizedFilter {
			if _, ok := normalizedPage[tag]; !ok {
				return false
			}
		}
		return true
	}
	for tag := range normalizedFilter {
		if _, ok := normalizedPage[tag]; ok {
			return true
		}
	}
	return false
}

func normalizeTagSet(tags []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		clean := strings.ToLower(strings.TrimSpace(tag))
		if clean == "" {
			continue
		}
		set[clean] = struct{}{}
	}
	return set
}

func matchQuery(page Page, query string, caseInsensitive bool) bool {
	terms := SplitQueryTerms(query)
	if len(terms) == 0 {
		return true
	}

	searchText := strings.Join([]string{
		page.Title,
		page.Section,
		strings.Join(page.Tags, " "),
		page.Content,
		page.Slug,
	}, " ")
	if caseInsensitive {
		searchText = strings.ToLower(searchText)
	}
	for _, term := range terms {
		target := term
		if !caseInsensitive {
			target = strings.TrimSpace(term)
		}
		if target == "" {
			continue
		}
		if !strings.Contains(searchText, target) {
			return false
		}
	}
	return true
}
