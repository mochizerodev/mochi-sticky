package wiki

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Index defines the navigation structure for wiki pages.
type Index struct {
	IndexVersion int            `yaml:"index_version"`
	Sections     []IndexSection `yaml:"sections"`
}

// IndexSection groups pages in the wiki index.
type IndexSection struct {
	Title string       `yaml:"title"`
	Slug  string       `yaml:"slug"`
	Order int          `yaml:"order"`
	Tags  []string     `yaml:"tags,omitempty"`
	Links SectionLinks `yaml:"links,omitempty"`
	Pages []string     `yaml:"pages"`
}

// ListPages expands the section's page slugs into fully qualified slugs.
func (s IndexSection) ListPages() []string {
	if strings.TrimSpace(s.Slug) == "" {
		return append([]string(nil), s.Pages...)
	}
	pages := make([]string, 0, len(s.Pages))
	prefix := strings.TrimSuffix(s.Slug, "/")
	for _, page := range s.Pages {
		pages = append(pages, prefix+"/"+page)
	}
	return pages
}

// RemoveSlug deletes a page slug from the index if present.
func (i *Index) RemoveSlug(slug string) bool {
	removed := false
	for sectionIndex, section := range i.Sections {
		if strings.TrimSpace(slug) == "" {
			continue
		}
		if strings.TrimSpace(section.Slug) == "" {
			newPages := make([]string, 0, len(section.Pages))
			for _, page := range section.Pages {
				if page == slug {
					removed = true
					continue
				}
				newPages = append(newPages, page)
			}
			i.Sections[sectionIndex].Pages = newPages
			continue
		}
		prefix := strings.TrimSuffix(section.Slug, "/") + "/"
		if !strings.HasPrefix(slug, prefix) {
			continue
		}
		relative := strings.TrimPrefix(slug, prefix)
		newPages := make([]string, 0, len(section.Pages))
		for _, page := range section.Pages {
			if page == relative {
				removed = true
				continue
			}
			newPages = append(newPages, page)
		}
		i.Sections[sectionIndex].Pages = newPages
	}
	return removed
}

// LoadIndex reads and parses a wiki index file.
func LoadIndex(path string) (Index, error) {
	return LoadIndexContext(context.Background(), path)
}

// LoadIndexContext reads and parses a wiki index file, honoring ctx cancellation.
func LoadIndexContext(ctx context.Context, path string) (Index, error) {
	select {
	case <-ctx.Done():
		return Index{}, ctx.Err()
	default:
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Index{}, ErrIndexNotFound
		}
		return Index{}, fmt.Errorf("wiki: failed to read index %s: %w", path, err)
	}
	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return Index{}, fmt.Errorf("wiki: failed to unmarshal index: %w", err)
	}
	select {
	case <-ctx.Done():
		return Index{}, ctx.Err()
	default:
	}
	return normalizeIndex(idx), nil
}

// SaveIndex writes the wiki index file to disk.
func SaveIndex(path string, index Index) error {
	return SaveIndexContext(context.Background(), path, index)
}

// SaveIndexContext writes the wiki index file to disk, honoring ctx cancellation.
func SaveIndexContext(ctx context.Context, path string, index Index) error {
	data, err := yaml.Marshal(normalizeIndex(index))
	if err != nil {
		return fmt.Errorf("wiki: failed to marshal index: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("wiki: failed to create index dir %s: %w", filepath.Dir(path), err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("wiki: failed to write index %s: %w", path, err)
	}
	return nil
}

func normalizeIndex(index Index) Index {
	if index.IndexVersion <= 0 {
		index.IndexVersion = 1
	}
	return index
}
