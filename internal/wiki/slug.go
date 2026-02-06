package wiki

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// Slugify converts a string into a URL-safe slug.
func Slugify(value string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			prevDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// NormalizeSlug validates and normalizes a slug to a safe relative path.
func NormalizeSlug(raw string) (string, error) {
	slug := strings.TrimSpace(raw)
	if slug == "" {
		return "", fmt.Errorf("slug is required")
	}
	clean := path.Clean(strings.ReplaceAll(slug, "\\", "/"))
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." || path.IsAbs(clean) {
		return "", fmt.Errorf("invalid slug: %s", slug)
	}
	return clean, nil
}

// SlugFromPath derives a slug from a page path relative to the wiki root.
func SlugFromPath(root, pagePath string) string {
	rel, err := filepath.Rel(root, pagePath)
	if err != nil {
		return ""
	}
	slug := strings.TrimSuffix(rel, ".md")
	return strings.ReplaceAll(slug, string(filepath.Separator), "/")
}
