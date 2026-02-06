package wiki

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// RootManifest groups manifest entries under a named root.
type RootManifest struct {
	Root   string
	Prefix string
	Pages  []ManifestEntry
}

// BuildRootManifest builds a manifest for a given wiki root.
func BuildRootManifest(root, prefix string) (RootManifest, error) {
	return BuildRootManifestContext(context.Background(), root, prefix)
}

// BuildRootManifestContext builds a manifest for a given wiki root, honoring ctx cancellation.
func BuildRootManifestContext(ctx context.Context, root, prefix string) (RootManifest, error) {
	indexPath := filepath.Join(root, "_index.yaml")
	index, indexErr := LoadIndexContext(ctx, indexPath)
	if indexErr != nil && !errors.Is(indexErr, ErrIndexNotFound) {
		return RootManifest{}, indexErr
	}

	var pages []Page
	var err error
	if indexErr == nil {
		pages, err = ListPagesFromIndexContext(ctx, root, index)
		if err != nil {
			return RootManifest{}, err
		}
	} else {
		pages, err = ListPagesContext(ctx, root)
		if err != nil {
			return RootManifest{}, err
		}
	}

	var manifest []ManifestEntry
	if indexErr == nil {
		manifest, err = BuildManifest(index, pages)
		if err != nil {
			return RootManifest{}, err
		}
	} else {
		manifest = BuildManifestFromPages(pages)
	}

	return RootManifest{
		Root:   root,
		Prefix: prefix,
		Pages:  manifest,
	}, nil
}

// FlattenManifests merges multiple root manifests with prefixed slugs.
func FlattenManifests(roots []RootManifest) []ManifestEntry {
	entries := make([]ManifestEntry, 0)
	seen := make(map[string]struct{})
	for _, root := range roots {
		for _, page := range root.Pages {
			slug := strings.TrimSpace(page.Slug)
			if strings.TrimSpace(root.Prefix) != "" {
				slug = strings.TrimSuffix(strings.TrimSpace(root.Prefix), "/") + "/" + slug
			}
			if _, exists := seen[slug]; exists {
				continue
			}
			seen[slug] = struct{}{}
			entries = append(entries, ManifestEntry{
				Title: page.Title,
				Slug:  slug,
				Order: page.Order,
			})
		}
	}
	return entries
}

// ValidateManifests checks for slug conflicts across roots.
func ValidateManifests(roots []RootManifest) error {
	seen := make(map[string]string)
	for _, root := range roots {
		for _, page := range root.Pages {
			slug := strings.TrimSpace(page.Slug)
			if strings.TrimSpace(root.Prefix) != "" {
				slug = strings.TrimSuffix(strings.TrimSpace(root.Prefix), "/") + "/" + slug
			}
			if prevRoot, exists := seen[slug]; exists {
				return fmt.Errorf("wiki: slug conflict %s between %s and %s", slug, prevRoot, root.Root)
			}
			seen[slug] = root.Root
		}
	}
	return nil
}
