package board

import "strings"

// ParseTags splits a comma-separated tag string into clean tags.
func ParseTags(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	return normalizeTags([]string{input})
}

// NormalizeTags cleans tag slices (trims, splits commas, removes empties, de-dupes).
func NormalizeTags(values []string) []string {
	return normalizeTags(values)
}

func normalizeTags(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, tag := range strings.Split(value, ",") {
			trimmed := strings.TrimSpace(tag)
			if trimmed == "" {
				continue
			}
			key := strings.ToLower(trimmed)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, trimmed)
		}
	}
	return out
}
