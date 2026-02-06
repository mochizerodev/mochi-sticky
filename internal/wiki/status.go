package wiki

import "strings"

// FilterPagesByStatus filters pages by status if a filter is provided.
func FilterPagesByStatus(pages []Page, status string) []Page {
	filter := strings.ToLower(strings.TrimSpace(status))
	if filter == "" {
		return pages
	}
	result := make([]Page, 0, len(pages))
	for _, page := range pages {
		if strings.ToLower(strings.TrimSpace(page.Status)) == filter {
			result = append(result, page)
		}
	}
	return result
}
