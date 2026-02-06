package templates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/storage"
)

// SeedDefaults copies bundled templates into the resolved template paths when empty.
// It also migrates legacy ADR/Wiki templates into the unified locations when needed.
func SeedDefaults(storageRoot string, paths storage.TemplatePaths) error {
	return SeedDefaultsContext(context.Background(), storageRoot, paths)
}

// SeedDefaultsContext copies bundled templates into the resolved template paths when empty,
// honoring ctx cancellation.
func SeedDefaultsContext(ctx context.Context, storageRoot string, paths storage.TemplatePaths) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if strings.TrimSpace(storageRoot) == "" {
		return fmt.Errorf("templates: storage root is required")
	}
	if err := ensureDir(paths.ADR); err != nil {
		return err
	}
	if err := ensureDir(paths.Task); err != nil {
		return err
	}
	if err := ensureDir(paths.Board); err != nil {
		return err
	}
	if err := ensureDir(paths.Wiki); err != nil {
		return err
	}

	legacyADR := filepath.Join(storageRoot, "adrs", "templates")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := copyDirIfEmpty(legacyADR, paths.ADR); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := copyEmbeddedDirIfEmpty("assets/adr", paths.ADR); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := copyEmbeddedDirIfEmpty("assets/task", paths.Task); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := copyEmbeddedDirIfEmpty("assets/board", paths.Board); err != nil {
		return err
	}
	legacyWiki := filepath.Join(storageRoot, "wiki", "templates")
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := copyDirIfEmpty(legacyWiki, paths.Wiki); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := copyEmbeddedDirIfEmpty("assets/wiki", paths.Wiki); err != nil {
		return err
	}

	if strings.TrimSpace(paths.WikiPDF) != "" {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := copyEmbeddedFileIfMissing("assets/wiki_pdf_template.tex", paths.WikiPDF); err != nil {
			return err
		}
	}

	return nil
}

func ensureDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("templates: expected directory but found file: %s", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("templates: failed to stat %s: %w", path, err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("templates: failed to create directory %s: %w", path, err)
	}
	return nil
}
