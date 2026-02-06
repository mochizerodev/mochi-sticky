package wiki

import "fmt"

// ValidateUniqueSlugs ensures no duplicate slugs exist in the provided pages.
func ValidateUniqueSlugs(pages []Page) error {
	seen := make(map[string]struct{}, len(pages))
	for _, page := range pages {
		if page.Slug == "" {
			continue
		}
		if _, exists := seen[page.Slug]; exists {
			return fmt.Errorf("wiki: %w: %s", ErrDuplicateSlug, page.Slug)
		}
		seen[page.Slug] = struct{}{}
	}
	return nil
}
