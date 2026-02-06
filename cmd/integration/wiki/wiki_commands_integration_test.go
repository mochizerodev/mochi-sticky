package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"mochi-sticky/internal/testutil"
	"mochi-sticky/internal/wiki"
)

func TestWikiCreateCommandCreatesPage(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)

	// Act
	slug := createWikiPage(
		t,
		repoRoot,
		storageRoot,
		"Getting Started",
		"--slug",
		"getting-started/intro",
		"--section",
		"Getting Started",
		"--order",
		"1",
		"--tags",
		"docs,setup",
		"--status",
		"draft",
	)

	// Assert
	path := filepath.Join(wikiRoot(storageRoot), filepath.FromSlash(slug)+".md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected wiki page to exist: %v", err)
	}
}

func TestWikiListCommandListsPages(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "List Page")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "list")
	if err != nil {
		t.Fatalf("wiki list: %v", err)
	}

	// Assert
	if !strings.Contains(out, slug+"\tList Page") {
		t.Fatalf("expected list output to contain page, got:\n%s", out)
	}
}

func TestWikiViewCommandPrintsContent(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "View Page")
	writeWikiContent(t, storageRoot, slug, "Hello wiki view.")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "view", slug)
	if err != nil {
		t.Fatalf("wiki view: %v", err)
	}

	// Assert
	if !strings.Contains(out, "Hello wiki view.") {
		t.Fatalf("expected view output to contain content, got:\n%s", out)
	}
}

func TestWikiEditCommandRunsEditor(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Edit Page")

	// Act
	editor := testutil.EditorCommandForTests()
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "edit", slug, "--editor", editor); err != nil {
		t.Fatalf("wiki edit: %v", err)
	}
}

func TestWikiSearchCommandFindsContent(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Search Page")
	writeWikiContent(t, storageRoot, slug, "UniqueSearchTokenAlpha")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "search", "UniqueSearchTokenAlpha")
	if err != nil {
		t.Fatalf("wiki search: %v", err)
	}

	// Assert
	if !strings.Contains(out, slug+":") {
		t.Fatalf("expected search output to contain slug, got:\n%s", out)
	}
}

func TestWikiListIncludeTemplatesCommandIncludesTemplates(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	writeTemplatePage(t, storageRoot, "template-page", "Template Page", "Template content")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "list", "--include-templates")
	if err != nil {
		t.Fatalf("wiki list include-templates: %v", err)
	}

	// Assert
	if !strings.Contains(out, "template-page\tTemplate Page") {
		t.Fatalf("expected template page in list, got:\n%s", out)
	}
}

func TestWikiSearchIncludeTemplatesCommandFindsTemplateContent(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	writeTemplatePage(t, storageRoot, "template-search", "Template Search", "UniqueTemplateToken")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "search", "UniqueTemplateToken", "--include-templates")
	if err != nil {
		t.Fatalf("wiki search include-templates: %v", err)
	}

	// Assert
	if !strings.Contains(out, "template-search:") {
		t.Fatalf("expected template search output to contain slug, got:\n%s", out)
	}
}

func TestWikiManifestCommandOutputsManifest(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Manifest Page")
	writeWikiContent(t, storageRoot, slug, "# Manifest\n\nBody content.")
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "index", "--write=true"); err != nil {
		t.Fatalf("wiki index: %v", err)
	}

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "manifest")
	if err != nil {
		t.Fatalf("wiki manifest: %v", err)
	}

	// Assert
	if !strings.Contains(out, "\"slug\":\""+slug+"\"") {
		t.Fatalf("expected manifest to include slug, got:\n%s", out)
	}
}

func TestWikiExportMarkdownCommandWritesFile(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Export Page")
	writeWikiContent(t, storageRoot, slug, "Exported markdown content.")
	outputPath := filepath.Join(t.TempDir(), "export.md")

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "export", "--format", "md", "--output", outputPath); err != nil {
		t.Fatalf("wiki export md: %v", err)
	}

	// Assert
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected export file: %v", err)
	}
	content := readFile(t, outputPath)
	if !strings.Contains(content, "Exported markdown content.") {
		t.Fatalf("expected exported content, got:\n%s", content)
	}
}

func TestWikiExportPDFCommandWritesFile(t *testing.T) {
	// Arrange
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not available")
	}
	if _, err := exec.LookPath("pdflatex"); err != nil {
		t.Skip("pdflatex not available")
	}
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "PDF Page")
	writeWikiContent(t, storageRoot, slug, "PDF export content.")
	outputPath := filepath.Join(t.TempDir(), "export.pdf")
	templatePath := filepath.Join(repoRoot, "internal", "templates", "assets", "wiki_pdf_template.tex")
	if _, err := os.Stat(templatePath); err != nil {
		t.Skipf("pdf template not found: %v", err)
	}

	// Act
	if _, err := runMochiSticky(
		t,
		repoRoot,
		storageRoot,
		"wiki",
		"export",
		"--format",
		"pdf",
		"--output",
		outputPath,
		"--title",
		"Export Title",
		"--author",
		"Export Author",
		"--template",
		templatePath,
	); err != nil {
		t.Fatalf("wiki export pdf: %v", err)
	}

	// Assert
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected pdf export file: %v", err)
	}
}

func TestWikiExportMarkdownWithRootCommandWritesFile(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	mainSlug := createWikiPage(t, repoRoot, storageRoot, "Main Page")
	writeWikiContent(t, storageRoot, mainSlug, "Main root content.")

	// On Windows, absolute paths include a drive prefix (e.g. C:\...) which conflicts with
	// our `path[:prefix]` parsing that splits on ":".
	// Create the external root under repoRoot and pass it as a relative path.
	externalRoot, err := os.MkdirTemp(repoRoot, "external-wiki-*")
	if err != nil {
		t.Fatalf("mkdir external root: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(externalRoot) })

	externalPath := filepath.Join(externalRoot, "external-page.md")
	if err := wiki.SavePage(externalPath, wiki.Page{
		Title:   "External Page",
		Slug:    "external-page",
		Status:  "published",
		Content: "External root content.",
	}); err != nil {
		t.Fatalf("save external page: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "export-multi.md")
	rootArg := filepath.Base(externalRoot)

	// Act
	if _, err := runMochiSticky(
		t,
		repoRoot,
		storageRoot,
		"wiki",
		"export",
		"--format",
		"md",
		"--root",
		rootArg+":ext",
		"--prefix",
		"main",
		"--output",
		outputPath,
	); err != nil {
		t.Fatalf("wiki export md root: %v", err)
	}

	// Assert
	content := readFile(t, outputPath)
	if !strings.Contains(content, "Main root content.") || !strings.Contains(content, "External root content.") {
		t.Fatalf("expected multi-root export content, got:\n%s", content)
	}
}

func TestWikiDeleteCommandRemovesPageAndIndex(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Delete Page")
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "index", "--write", "true"); err != nil {
		t.Fatalf("wiki index: %v", err)
	}

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "delete", slug, "--update-index"); err != nil {
		t.Fatalf("wiki delete: %v", err)
	}

	// Assert
	path := filepath.Join(wikiRoot(storageRoot), filepath.FromSlash(slug)+".md")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected page to be deleted, got: %v", err)
	}
	indexPath := filepath.Join(wikiRoot(storageRoot), "_index.yaml")
	indexContent := readFile(t, indexPath)
	if strings.Contains(indexContent, slug) {
		t.Fatalf("expected index to remove slug, got:\n%s", indexContent)
	}
}

func TestWikiIndexCommandOutputsJSON(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Index Page")
	writeTemplatePage(t, storageRoot, "template-index", "Template Index", "Template index content")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "index", "--include-templates", "--write=false")
	if err != nil {
		t.Fatalf("wiki index: %v", err)
	}

	// Assert
	if !strings.Contains(out, slug) {
		t.Fatalf("expected index output to include slug, got:\n%s", out)
	}
	if !strings.Contains(out, "template-index") {
		t.Fatalf("expected index output to include template slug, got:\n%s", out)
	}
}

func TestWikiListStatusCommandFilters(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	draftSlug := createWikiPage(t, repoRoot, storageRoot, "Draft Page", "--status", "draft")
	_ = createWikiPage(t, repoRoot, storageRoot, "Published Page", "--status", "published")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "list", "--status", "draft")
	if err != nil {
		t.Fatalf("wiki list status: %v", err)
	}

	// Assert
	if !strings.Contains(out, draftSlug) {
		t.Fatalf("expected draft page in list, got:\n%s", out)
	}
	if strings.Contains(out, "Published Page") {
		t.Fatalf("expected published page to be filtered out, got:\n%s", out)
	}
}

func TestWikiSearchStatusCommandFilters(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Status Search", "--status", "draft")
	writeWikiContent(t, storageRoot, slug, "StatusSearchToken")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "search", "StatusSearchToken", "--status", "draft")
	if err != nil {
		t.Fatalf("wiki search status: %v", err)
	}

	// Assert
	if !strings.Contains(out, slug+":") {
		t.Fatalf("expected status search output to contain slug, got:\n%s", out)
	}
}

func TestWikiLintCommandReportsNoIssues(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupWikiStorage(t)
	slug := createWikiPage(t, repoRoot, storageRoot, "Lint Page")
	writeWikiContent(t, storageRoot, slug, "# Lint Page\n\nClean content.")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "wiki", "lint")
	if err != nil {
		t.Fatalf("wiki lint: %v", err)
	}

	// Assert
	if !strings.Contains(out, "No issues found.") {
		t.Fatalf("expected lint to report no issues, got:\n%s", out)
	}
}
