package wiki

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExportMarkdown compiles pages into a single Markdown document.
func ExportMarkdown(root string, manifest []ManifestEntry) ([]byte, error) {
	return ExportMarkdownContext(context.Background(), root, manifest)
}

// ExportMarkdownContext compiles pages into a single Markdown document and honors ctx cancellation.
func ExportMarkdownContext(ctx context.Context, root string, manifest []ManifestEntry) ([]byte, error) {
	var buf bytes.Buffer
	for i, entry := range manifest {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if strings.TrimSpace(entry.Slug) == "" {
			return nil, fmt.Errorf("wiki: %w", ErrPageNotFound)
		}
		pagePath := filepath.Join(root, filepath.FromSlash(entry.Slug)+".md")
		page, err := LoadPage(pagePath)
		if err != nil {
			return nil, err
		}
		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = page.Title
		}
		if title != "" {
			buf.WriteString("# ")
			buf.WriteString(title)
			buf.WriteString("\n\n")
		}
		if strings.TrimSpace(page.Content) != "" {
			buf.WriteString(strings.TrimSpace(page.Content))
			buf.WriteString("\n")
		}
		if i < len(manifest)-1 {
			buf.WriteString("\n---\n\n")
		}
	}
	return buf.Bytes(), nil
}

type rootEntry struct {
	Root       string
	Title      string
	Slug       string
	SourceSlug string
	Order      int
}

// ExportMarkdownMulti compiles pages from multiple roots.
func ExportMarkdownMulti(rootManifests []RootManifest) ([]byte, error) {
	return ExportMarkdownMultiContext(context.Background(), rootManifests)
}

// ExportMarkdownMultiContext compiles pages from multiple roots and honors ctx cancellation.
func ExportMarkdownMultiContext(ctx context.Context, rootManifests []RootManifest) ([]byte, error) {
	entries := flattenRootEntries(rootManifests)
	var buf bytes.Buffer
	for i, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if strings.TrimSpace(entry.SourceSlug) == "" {
			return nil, fmt.Errorf("wiki: %w", ErrPageNotFound)
		}
		pagePath := filepath.Join(entry.Root, filepath.FromSlash(entry.SourceSlug)+".md")
		page, err := LoadPage(pagePath)
		if err != nil {
			return nil, err
		}
		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = page.Title
		}
		if title != "" {
			buf.WriteString("# ")
			buf.WriteString(title)
			buf.WriteString("\n\n")
		}
		if strings.TrimSpace(page.Content) != "" {
			buf.WriteString(strings.TrimSpace(page.Content))
			buf.WriteString("\n")
		}
		if i < len(entries)-1 {
			buf.WriteString("\n---\n\n")
		}
	}
	return buf.Bytes(), nil
}

func flattenRootEntries(roots []RootManifest) []rootEntry {
	entries := make([]rootEntry, 0)
	for _, root := range roots {
		for _, page := range root.Pages {
			slug := strings.TrimSpace(page.Slug)
			if strings.TrimSpace(root.Prefix) != "" {
				slug = strings.TrimSuffix(strings.TrimSpace(root.Prefix), "/") + "/" + slug
			}
			entries = append(entries, rootEntry{
				Root:       root.Root,
				Title:      page.Title,
				Slug:       slug,
				SourceSlug: page.Slug,
				Order:      page.Order,
			})
		}
	}
	return entries
}

// WriteExport writes exported content to disk.
func WriteExport(path string, data []byte) error {
	return WriteExportContext(context.Background(), path, data)
}

// WriteExportContext writes exported content to disk, honoring ctx cancellation.
func WriteExportContext(ctx context.Context, path string, data []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("wiki: failed to create export dir %s: %w", filepath.Dir(path), err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("wiki: failed to write export %s: %w", path, err)
	}
	return nil
}
