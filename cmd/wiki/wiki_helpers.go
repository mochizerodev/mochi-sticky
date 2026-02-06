package wiki

import (
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/wiki"
)

func wikiRoot(storageRoot string) string {
	return filepath.Join(storageRoot, "wiki")
}

func pagePath(root, slug string) (string, error) {
	clean, err := wiki.NormalizeSlug(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, filepath.FromSlash(clean)+".md"), nil
}

func slugFromPath(root, pagePath string) string {
	rel, err := filepath.Rel(root, pagePath)
	if err != nil {
		return ""
	}
	slug := strings.TrimSuffix(rel, ".md")
	return strings.ReplaceAll(slug, string(os.PathSeparator), "/")
}
