package wiki

import (
	"fmt"
	"sort"
	"strings"
)

// SectionLinks captures relationships between sections.
type SectionLinks struct {
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	RelatedTo []string `yaml:"related_to,omitempty" json:"related_to,omitempty"`
}

// FindSection resolves a section by slug or title and returns it with its index.
func FindSection(index Index, key string) (IndexSection, int, error) {
	needle := strings.TrimSpace(key)
	if needle == "" {
		empty := make([]int, 0)
		for i, candidate := range index.Sections {
			if strings.TrimSpace(candidate.Slug) == "" {
				empty = append(empty, i)
			}
		}
		if len(empty) == 1 {
			return index.Sections[empty[0]], empty[0], nil
		}
		if len(empty) > 1 {
			return IndexSection{}, -1, fmt.Errorf("wiki: multiple sections with empty slug")
		}
		return IndexSection{}, -1, ErrSectionNotFound
	}

	needleLower := strings.ToLower(needle)
	for i, candidate := range index.Sections {
		if strings.ToLower(strings.TrimSpace(candidate.Slug)) == needleLower {
			return candidate, i, nil
		}
	}

	if normalized, err := NormalizeSlug(needle); err == nil {
		normalizedLower := strings.ToLower(strings.TrimSpace(normalized))
		for i, candidate := range index.Sections {
			if strings.ToLower(strings.TrimSpace(candidate.Slug)) == normalizedLower {
				return candidate, i, nil
			}
		}
	}

	var matchIndex = -1
	for i, candidate := range index.Sections {
		if strings.ToLower(strings.TrimSpace(candidate.Title)) == needleLower {
			if matchIndex != -1 {
				return IndexSection{}, -1, fmt.Errorf("wiki: multiple sections match %s", needle)
			}
			matchIndex = i
		}
	}
	if matchIndex != -1 {
		return index.Sections[matchIndex], matchIndex, nil
	}
	return IndexSection{}, -1, ErrSectionNotFound
}

// ResolveLinkedSections returns linked sections based on the link types.
func ResolveLinkedSections(index Index, section IndexSection, linkTypes []string) ([]IndexSection, error) {
	types := normalizeLinkTypes(linkTypes)
	seen := make(map[string]struct{})
	sections := make([]IndexSection, 0)
	add := func(target string) error {
		if strings.TrimSpace(target) == "" {
			return nil
		}
		resolved, _, err := FindSection(index, target)
		if err != nil {
			return err
		}
		key := sectionKey(resolved)
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		sections = append(sections, resolved)
		return nil
	}

	for _, linkType := range types {
		switch linkType {
		case "depends_on":
			for _, target := range section.Links.DependsOn {
				if err := add(target); err != nil {
					return nil, err
				}
			}
		case "related_to":
			for _, target := range section.Links.RelatedTo {
				if err := add(target); err != nil {
					return nil, err
				}
			}
		}
	}
	return sections, nil
}

// BuildManifestForSections builds a manifest for a subset of sections.
func BuildManifestForSections(index Index, pages []Page, sections []IndexSection) ([]ManifestEntry, error) {
	pageBySlug := make(map[string]Page, len(pages))
	for _, page := range pages {
		slug := strings.TrimSpace(page.Slug)
		if slug == "" {
			continue
		}
		pageBySlug[slug] = page
	}

	manifest := make([]ManifestEntry, 0)
	seen := make(map[string]struct{})
	for _, section := range sections {
		for _, fullSlug := range section.ListPages() {
			if _, ok := seen[fullSlug]; ok {
				continue
			}
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
			seen[fullSlug] = struct{}{}
		}
	}
	return manifest, nil
}

type SectionFilterOptions struct {
	Tags       []string
	TagMode    string
	LinkType   string
	LinkTarget string
}

// FilterSections filters sections by tags and links.
func FilterSections(sections []IndexSection, opts SectionFilterOptions) []IndexSection {
	if len(opts.Tags) == 0 && strings.TrimSpace(opts.LinkType) == "" && strings.TrimSpace(opts.LinkTarget) == "" {
		return sections
	}
	result := make([]IndexSection, 0, len(sections))
	for _, section := range sections {
		if !matchesSectionTags(section, opts.Tags, opts.TagMode) {
			continue
		}
		if !matchesSectionLinks(section, opts.LinkType, opts.LinkTarget) {
			continue
		}
		result = append(result, section)
	}
	return result
}

func matchesSectionTags(section IndexSection, tags []string, mode string) bool {
	if len(tags) == 0 {
		return true
	}
	return matchTags(section.Tags, tags, mode)
}

func matchesSectionLinks(section IndexSection, linkType, linkTarget string) bool {
	linkType = strings.ToLower(strings.TrimSpace(linkType))
	linkTarget = strings.ToLower(strings.TrimSpace(linkTarget))
	if linkType == "" && linkTarget == "" {
		return true
	}
	links := sectionLinksByType(section)
	if linkType != "" {
		targets := links[linkType]
		if len(targets) == 0 {
			return false
		}
		if linkTarget == "" {
			return true
		}
		for _, target := range targets {
			if strings.ToLower(strings.TrimSpace(target)) == linkTarget {
				return true
			}
		}
		return false
	}
	if linkTarget == "" {
		return false
	}
	for _, targets := range links {
		for _, target := range targets {
			if strings.ToLower(strings.TrimSpace(target)) == linkTarget {
				return true
			}
		}
	}
	return false
}

func sectionLinksByType(section IndexSection) map[string][]string {
	result := map[string][]string{
		"depends_on": section.Links.DependsOn,
		"related_to": section.Links.RelatedTo,
	}
	return result
}

func normalizeLinkTypes(types []string) []string {
	if len(types) == 0 {
		return []string{"depends_on", "related_to"}
	}
	clean := make([]string, 0, len(types))
	seen := make(map[string]struct{})
	for _, linkType := range types {
		value := strings.ToLower(strings.TrimSpace(linkType))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		clean = append(clean, value)
	}
	sort.Strings(clean)
	return clean
}

func sectionKey(section IndexSection) string {
	key := strings.TrimSpace(section.Slug)
	if key == "" {
		key = strings.TrimSpace(section.Title)
	}
	return strings.ToLower(key)
}
