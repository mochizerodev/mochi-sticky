package wiki

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/shared"
)

// ListPages loads all wiki pages from the root directory.
func ListPages(root string) ([]Page, error) {
	return ListPagesContext(context.Background(), root)
}

// ListPagesContext loads all wiki pages from the root directory, honoring ctx cancellation.
func ListPagesContext(ctx context.Context, root string) ([]Page, error) {
	return listPagesContext(ctx, root, false)
}

// ListPagesWithTemplates includes template pages when set to true.
func ListPagesWithTemplates(root string, includeTemplates bool) ([]Page, error) {
	return ListPagesWithTemplatesRoot(root, includeTemplates, "")
}

// ListPagesWithTemplatesRoot includes template pages from an optional templates root.
func ListPagesWithTemplatesRoot(root string, includeTemplates bool, templatesRoot string) ([]Page, error) {
	return ListPagesWithTemplatesRootContext(context.Background(), root, includeTemplates, templatesRoot)
}

// ListPagesWithTemplatesRootContext includes template pages from an optional templates root, honoring ctx cancellation.
func ListPagesWithTemplatesRootContext(ctx context.Context, root string, includeTemplates bool, templatesRoot string) ([]Page, error) {
	pages, err := listPagesContext(ctx, root, includeTemplates)
	if err != nil {
		return nil, err
	}
	if !includeTemplates {
		return pages, nil
	}
	if strings.TrimSpace(templatesRoot) == "" || shared.IsSubpath(root, templatesRoot) {
		return pages, nil
	}

	templatePages, err := listPagesContext(ctx, templatesRoot, true)
	if err != nil {
		return nil, err
	}
	pages = append(pages, templatePages...)
	if err := ValidateUniqueSlugs(pages); err != nil {
		return nil, err
	}
	return pages, nil
}

func listPagesContext(ctx context.Context, root string, includeTemplates bool) ([]Page, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("wiki: failed to stat wiki root %s: %w", root, err)
	}

	var pages []Page
	err := filepath.WalkDir(root, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if d.IsDir() {
			if !includeTemplates && filepath.Base(entryPath) == "templates" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entryPath) != ".md" {
			return nil
		}
		if filepath.Base(entryPath) == "_index.yaml" {
			return nil
		}
		page, err := LoadPage(entryPath)
		if err != nil {
			return err
		}
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("wiki: failed to list pages: %w", err)
	}
	if err := ValidateUniqueSlugs(pages); err != nil {
		return nil, err
	}
	return pages, nil
}

// ListPagesFromIndex loads pages in the order defined by the index.
func ListPagesFromIndex(root string, index Index) ([]Page, error) {
	return ListPagesFromIndexContext(context.Background(), root, index)
}

// ListPagesFromIndexContext loads pages in the order defined by the index, honoring ctx cancellation.
func ListPagesFromIndexContext(ctx context.Context, root string, index Index) ([]Page, error) {
	pages := make([]Page, 0)
	for _, section := range index.Sections {
		for _, pageSlug := range section.Pages {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			fullSlug := pageSlug
			if strings.TrimSpace(section.Slug) != "" {
				fullSlug = path.Join(section.Slug, pageSlug)
			}
			pagePath := filepath.Join(root, filepath.FromSlash(fullSlug)+".md")
			page, err := LoadPage(pagePath)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(page.Slug) == "" {
				page.Slug = fullSlug
			}
			pages = append(pages, page)
		}
	}
	if err := ValidateUniqueSlugs(pages); err != nil {
		return nil, err
	}
	return pages, nil
}
