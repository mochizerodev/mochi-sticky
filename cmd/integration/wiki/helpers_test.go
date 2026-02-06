package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mochi-sticky/internal/testutil"
	"mochi-sticky/internal/wiki"
)

func setupWikiStorage(t *testing.T) (repoRoot, storageRoot string) {
	return testutil.SetupStorage(t)
}

func runMochiSticky(t *testing.T, repoRoot, storageRoot string, args ...string) (string, error) {
	return testutil.RunMochiSticky(t, repoRoot, storageRoot, args...)
}

func wikiRoot(storageRoot string) string {
	return filepath.Join(storageRoot, "wiki")
}

func templateRoot(storageRoot string) string {
	return filepath.Join(storageRoot, "templates", "wiki")
}

func createWikiPage(t *testing.T, repoRoot, storageRoot, title string, args ...string) string {
	t.Helper()

	cmdArgs := append([]string{"wiki", "create", title}, args...)
	out, err := runMochiSticky(t, repoRoot, storageRoot, cmdArgs...)
	if err != nil {
		t.Fatalf("wiki create: %v", err)
	}
	return parseCreatedWikiSlug(t, out)
}

func parseCreatedWikiSlug(t *testing.T, output string) string {
	t.Helper()

	trimmed := strings.TrimSpace(output)
	prefix := "Created wiki page "
	if !strings.HasPrefix(trimmed, prefix) {
		t.Fatalf("unexpected create output: %q", trimmed)
	}
	slug := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	if slug == "" {
		t.Fatalf("expected wiki slug in output: %q", trimmed)
	}
	return slug
}

func writeWikiContent(t *testing.T, storageRoot, slug, content string) {
	t.Helper()

	path := filepath.Join(wikiRoot(storageRoot), filepath.FromSlash(slug)+".md")
	page, err := wiki.LoadPage(path)
	if err != nil {
		t.Fatalf("load wiki page: %v", err)
	}
	page.Content = content
	if err := wiki.SavePage(path, page); err != nil {
		t.Fatalf("save wiki page: %v", err)
	}
}

func writeTemplatePage(t *testing.T, storageRoot, slug, title, content string) {
	t.Helper()

	path := filepath.Join(templateRoot(storageRoot), filepath.FromSlash(slug)+".md")
	page := wiki.Page{
		Title:   title,
		Slug:    slug,
		Status:  "published",
		Content: content,
	}
	if err := wiki.SavePage(path, page); err != nil {
		t.Fatalf("save template page: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(data)
}
