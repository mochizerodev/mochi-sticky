package integration

import (
	"path/filepath"
	"strings"
	"testing"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/testutil"
)

func setupADRStorage(t *testing.T) (repoRoot, storageRoot string) {
	return testutil.SetupStorage(t)
}

func runMochiSticky(t *testing.T, repoRoot, storageRoot string, args ...string) (string, error) {
	return testutil.RunMochiSticky(t, repoRoot, storageRoot, args...)
}

func runMochiStickyWithInput(t *testing.T, repoRoot, storageRoot, input string, args ...string) (string, error) {
	return testutil.RunMochiStickyWithInput(t, repoRoot, storageRoot, input, args...)
}

func adrRoot(storageRoot string) string {
	return filepath.Join(storageRoot, "adrs")
}

func createADR(t *testing.T, repoRoot, storageRoot, title string, args ...string) string {
	t.Helper()

	cmdArgs := append([]string{"adr", "create", title}, args...)
	out, err := runMochiSticky(t, repoRoot, storageRoot, cmdArgs...)
	if err != nil {
		t.Fatalf("adr create: %v", err)
	}
	return parseCreatedADRID(t, out)
}

func createADRWithBody(t *testing.T, repoRoot, storageRoot, title, body string, args ...string) string {
	t.Helper()

	cmdArgs := append([]string{"adr", "create", title}, args...)
	out, err := runMochiStickyWithInput(t, repoRoot, storageRoot, body, cmdArgs...)
	if err != nil {
		t.Fatalf("adr create: %v", err)
	}
	return parseCreatedADRID(t, out)
}

func parseCreatedADRID(t *testing.T, output string) string {
	t.Helper()

	trimmed := strings.TrimSpace(output)
	prefix := "Created ADR "
	if !strings.HasPrefix(trimmed, prefix) {
		t.Fatalf("unexpected create output: %q", trimmed)
	}
	id := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	if id == "" {
		t.Fatalf("expected ADR ID in output: %q", trimmed)
	}
	return id
}

func loadADR(t *testing.T, storageRoot, id string) adr.ADR {
	t.Helper()

	parsed, err := adr.ParseID(id)
	if err != nil {
		t.Fatalf("parse adr id: %v", err)
	}
	repo, err := adr.NewRepository(adrRoot(storageRoot))
	if err != nil {
		t.Fatalf("new adr repo: %v", err)
	}
	record, err := repo.GetADRByID(parsed)
	if err != nil {
		t.Fatalf("get adr: %v", err)
	}
	return record
}
