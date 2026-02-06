package integration

import (
	"strings"
	"testing"

	"mochi-sticky/internal/testutil"
)

func TestADRCreateCommandCreatesRecord(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupADRStorage(t)
	body := "## Context\n\nInput body\n\n## Decision\n\nDecide\n\n## Consequences\n\nDone\n"

	// Act
	id := createADRWithBody(
		t,
		repoRoot,
		storageRoot,
		"ADR Title",
		body,
		"--status",
		"proposed",
		"--date",
		"2022-01-02",
		"--tags",
		"alpha,beta",
		"--links",
		"task-1,wiki/slug",
		"--body",
		"-",
	)

	// Assert
	record := loadADR(t, storageRoot, id)
	if record.Title != "ADR Title" {
		t.Fatalf("expected title %q, got %q", "ADR Title", record.Title)
	}
	if record.Status != "proposed" {
		t.Fatalf("expected status %q, got %q", "proposed", record.Status)
	}
	if record.Date.Format("2006-01-02") != "2022-01-02" {
		t.Fatalf("expected date %q, got %q", "2022-01-02", record.Date.Format("2006-01-02"))
	}
	if !strings.Contains(record.Content, "## Context") {
		t.Fatalf("expected ADR content to include headings")
	}
}

func TestADRListCommandFilters(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupADRStorage(t)
	idOne := createADR(t, repoRoot, storageRoot, "Alpha Decision", "--status", "accepted", "--date", "2020-01-01", "--tags", "alpha,core")
	idTwo := createADR(t, repoRoot, storageRoot, "Beta Decision", "--status", "proposed", "--date", "2023-01-01", "--tags", "beta")

	// Act
	statusOut, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "list", "--status", "accepted")
	if err != nil {
		t.Fatalf("adr list status: %v", err)
	}
	tagOut, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "list", "--tags", "beta")
	if err != nil {
		t.Fatalf("adr list tags: %v", err)
	}
	queryOut, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "list", "--query", "Alpha")
	if err != nil {
		t.Fatalf("adr list query: %v", err)
	}
	sinceOut, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "list", "--since", "2022-01-01")
	if err != nil {
		t.Fatalf("adr list since: %v", err)
	}
	untilOut, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "list", "--until", "2021-12-31")
	if err != nil {
		t.Fatalf("adr list until: %v", err)
	}

	// Assert
	if !strings.Contains(statusOut, idOne) || strings.Contains(statusOut, idTwo) {
		t.Fatalf("expected status filter to include %s only, got:\n%s", idOne, statusOut)
	}
	if !strings.Contains(tagOut, idTwo) || strings.Contains(tagOut, idOne) {
		t.Fatalf("expected tag filter to include %s only, got:\n%s", idTwo, tagOut)
	}
	if !strings.Contains(queryOut, idOne) || strings.Contains(queryOut, idTwo) {
		t.Fatalf("expected query filter to include %s only, got:\n%s", idOne, queryOut)
	}
	if !strings.Contains(sinceOut, idTwo) || strings.Contains(sinceOut, idOne) {
		t.Fatalf("expected since filter to include %s only, got:\n%s", idTwo, sinceOut)
	}
	if !strings.Contains(untilOut, idOne) || strings.Contains(untilOut, idTwo) {
		t.Fatalf("expected until filter to include %s only, got:\n%s", idOne, untilOut)
	}
}

func TestADRViewCommandPrintsFile(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupADRStorage(t)
	id := createADR(t, repoRoot, storageRoot, "View ADR")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "view", id)
	if err != nil {
		t.Fatalf("adr view: %v", err)
	}

	// Assert
	if !strings.Contains(out, "title: View ADR") {
		t.Fatalf("expected ADR content in view output, got:\n%s", out)
	}
}

func TestADREditCommandRunsEditor(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupADRStorage(t)
	id := createADR(t, repoRoot, storageRoot, "Edit ADR")

	// Act
	editor := testutil.EditorCommandForTests()
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "edit", id, "--editor", editor); err != nil {
		t.Fatalf("adr edit: %v", err)
	}
}

func TestADRMoveCommandUpdatesStatus(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupADRStorage(t)
	id := createADR(t, repoRoot, storageRoot, "Move ADR", "--status", "proposed")

	// Act
	if _, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "move", id, "accepted"); err != nil {
		t.Fatalf("adr move: %v", err)
	}

	// Assert
	record := loadADR(t, storageRoot, id)
	if record.Status != "accepted" {
		t.Fatalf("expected status %q, got %q", "accepted", record.Status)
	}
}

func TestADRStatusesCommandListsDefaults(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupADRStorage(t)

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "statuses")
	if err != nil {
		t.Fatalf("adr statuses: %v", err)
	}

	// Assert
	for _, status := range []string{"proposed", "accepted", "rejected", "deprecated", "superseded"} {
		if !strings.Contains(out, status) {
			t.Fatalf("expected status %q in output, got:\n%s", status, out)
		}
	}
}

func TestADRLintCommandReportsOK(t *testing.T) {
	// Arrange
	repoRoot, storageRoot := setupADRStorage(t)
	_ = createADR(t, repoRoot, storageRoot, "Lint ADR")

	// Act
	out, err := runMochiSticky(t, repoRoot, storageRoot, "adr", "lint")
	if err != nil {
		t.Fatalf("adr lint: %v", err)
	}

	// Assert
	if !strings.Contains(out, "OK") {
		t.Fatalf("expected lint to report OK, got:\n%s", out)
	}
}
