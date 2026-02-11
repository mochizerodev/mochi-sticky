package integration

import (
	"testing"

	"mochi-sticky/internal/testutil"
)

func setupReleaseStorage(t *testing.T) (repoRoot, storageRoot string) {
	return testutil.SetupStorage(t)
}

func runMochiSticky(t *testing.T, repoRoot, storageRoot string, args ...string) (string, error) {
	return testutil.RunMochiSticky(t, repoRoot, storageRoot, args...)
}
