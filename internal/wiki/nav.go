package wiki

import (
	"context"
	"fmt"
	"strings"
)

// NavNode represents a node in the wiki navigation tree.
type NavNode struct {
	Title string       `json:"title"`
	Slug  string       `json:"slug"`
	Order int          `json:"order"`
	Tags  []string     `json:"tags,omitempty"`
	Links SectionLinks `json:"links,omitempty"`
	Pages []PageRef    `json:"pages"`
}

// PageRef references a page in the navigation tree.
type PageRef struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
	Order int    `json:"order"`
}

// BuildNavTree builds a deterministic navigation tree from the index and pages.
func BuildNavTree(index Index, pages []Page) ([]NavNode, error) {
	return BuildNavTreeContext(context.Background(), index, pages)
}

// BuildNavTreeContext builds a deterministic navigation tree from the index and pages, honoring ctx cancellation.
func BuildNavTreeContext(ctx context.Context, index Index, pages []Page) ([]NavNode, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	pageBySlug := make(map[string]Page, len(pages))
	for _, page := range pages {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if strings.TrimSpace(page.Slug) == "" {
			continue
		}
		if _, exists := pageBySlug[page.Slug]; exists {
			return nil, fmt.Errorf("wiki: %w: %s", ErrDuplicateSlug, page.Slug)
		}
		pageBySlug[page.Slug] = page
	}

	nodes := make([]NavNode, 0, len(index.Sections))
	for _, section := range index.Sections {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		node := NavNode{
			Title: section.Title,
			Slug:  section.Slug,
			Order: section.Order,
			Tags:  section.Tags,
			Links: section.Links,
		}
		for _, fullSlug := range section.ListPages() {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			page, ok := pageBySlug[fullSlug]
			if !ok {
				return nil, fmt.Errorf("wiki: missing page for slug %s", fullSlug)
			}
			node.Pages = append(node.Pages, PageRef{
				Title: page.Title,
				Slug:  fullSlug,
				Order: page.Order,
			})
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
