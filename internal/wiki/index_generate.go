package wiki

import (
	"context"
	"sort"
	"strings"
)

type pageEntry struct {
	slug  string
	title string
	order int
}

type sectionAccum struct {
	section IndexSection
	pages   []pageEntry
}

// GenerateIndex builds an index from the provided pages.
func GenerateIndex(pages []Page) (Index, error) {
	return GenerateIndexContext(context.Background(), pages)
}

// GenerateIndexContext builds an index from the provided pages, honoring ctx cancellation.
func GenerateIndexContext(ctx context.Context, pages []Page) (Index, error) {
	select {
	case <-ctx.Done():
		return Index{}, ctx.Err()
	default:
	}
	if err := ValidateUniqueSlugs(pages); err != nil {
		return Index{}, err
	}
	sections := make(map[string]*sectionAccum)
	for _, page := range pages {
		slug := strings.TrimSpace(page.Slug)
		if slug == "" {
			continue
		}
		sectionSlug, pageSlug := splitSlug(slug)
		sectionTitle := strings.TrimSpace(page.Section)
		if sectionTitle == "" {
			if sectionSlug == "" {
				sectionTitle = "General"
			} else {
				sectionTitle = titleFromSlug(sectionSlug)
			}
		}
		acc := sections[sectionSlug]
		if acc == nil {
			acc = &sectionAccum{
				section: IndexSection{
					Title: sectionTitle,
					Slug:  sectionSlug,
				},
			}
			sections[sectionSlug] = acc
		}
		acc.pages = append(acc.pages, pageEntry{
			slug:  pageSlug,
			title: page.Title,
			order: page.Order,
		})
	}

	sectionList := make([]*sectionAccum, 0, len(sections))
	for _, acc := range sections {
		select {
		case <-ctx.Done():
			return Index{}, ctx.Err()
		default:
		}
		sectionList = append(sectionList, acc)
	}
	sort.Slice(sectionList, func(i, j int) bool {
		if sectionList[i].section.Slug == "" && sectionList[j].section.Slug != "" {
			return true
		}
		if sectionList[i].section.Slug != "" && sectionList[j].section.Slug == "" {
			return false
		}
		left := strings.ToLower(sectionList[i].section.Title)
		right := strings.ToLower(sectionList[j].section.Title)
		if left != right {
			return left < right
		}
		return sectionList[i].section.Slug < sectionList[j].section.Slug
	})

	index := Index{Sections: make([]IndexSection, 0, len(sectionList))}
	for _, acc := range sectionList {
		select {
		case <-ctx.Done():
			return Index{}, ctx.Err()
		default:
		}
		sort.Slice(acc.pages, func(i, j int) bool {
			left := acc.pages[i]
			right := acc.pages[j]
			if left.order != 0 || right.order != 0 {
				if left.order == 0 {
					return false
				}
				if right.order == 0 {
					return true
				}
				if left.order != right.order {
					return left.order < right.order
				}
			}
			leftTitle := strings.ToLower(left.title)
			rightTitle := strings.ToLower(right.title)
			if leftTitle != rightTitle {
				return leftTitle < rightTitle
			}
			return left.slug < right.slug
		})
		section := acc.section
		section.Pages = make([]string, 0, len(acc.pages))
		for _, page := range acc.pages {
			select {
			case <-ctx.Done():
				return Index{}, ctx.Err()
			default:
			}
			section.Pages = append(section.Pages, page.slug)
		}
		index.Sections = append(index.Sections, section)
	}
	return normalizeIndex(index), nil
}

func splitSlug(slug string) (sectionSlug, pageSlug string) {
	parts := strings.SplitN(slug, "/", 2)
	if len(parts) == 1 {
		return "", slug
	}
	return parts[0], parts[1]
}

func titleFromSlug(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool {
		return r == '-' || r == '_' || r == '/'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
