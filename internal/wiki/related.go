package wiki

import (
	"sort"
	"strings"
)

// RelatedPage describes a related wiki page match.
type RelatedPage struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	SharedTags  []string `json:"shared_tags"`
	SharedCount int      `json:"shared_count"`
}

// RelatedPages finds related pages based on shared tags.
func RelatedPages(pages []Page, slug string) ([]RelatedPage, error) {
	targetSlug := strings.TrimSpace(slug)
	if targetSlug == "" {
		return nil, ErrPageNotFound
	}
	bySlug := make(map[string]Page, len(pages))
	for _, page := range pages {
		if strings.TrimSpace(page.Slug) == "" {
			continue
		}
		bySlug[page.Slug] = page
	}
	target, ok := bySlug[targetSlug]
	if !ok {
		return nil, ErrPageNotFound
	}

	targetTags := normalizeTags(target.Tags)
	if len(targetTags) == 0 {
		return nil, nil
	}

	results := make([]RelatedPage, 0)
	for _, page := range pages {
		if page.Slug == targetSlug {
			continue
		}
		if shouldSkipStatus(page.Status) {
			continue
		}
		shared := sharedTags(targetTags, normalizeTags(page.Tags))
		if len(shared) == 0 {
			continue
		}
		results = append(results, RelatedPage{
			Slug:        page.Slug,
			Title:       page.Title,
			SharedTags:  shared,
			SharedCount: len(shared),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].SharedCount != results[j].SharedCount {
			return results[i].SharedCount > results[j].SharedCount
		}
		return strings.ToLower(results[i].Title) < strings.ToLower(results[j].Title)
	})
	return results, nil
}

func shouldSkipStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft", "archived":
		return true
	default:
		return false
	}
}

func normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		clean := strings.TrimSpace(strings.ToLower(tag))
		if clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}
	sort.Strings(result)
	return result
}

func sharedTags(a, b []string) []string {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(a))
	for _, tag := range a {
		set[tag] = struct{}{}
	}
	shared := make([]string, 0)
	for _, tag := range b {
		if _, ok := set[tag]; ok {
			shared = append(shared, tag)
		}
	}
	sort.Strings(shared)
	return shared
}
