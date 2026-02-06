package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// FindRepoRoot walks up from the current working directory to find go.mod.
func FindRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}

// SetupStorage initializes a temp storage root and runs `mochi-sticky init`.
func SetupStorage(t *testing.T) (repoRoot, storageRoot string) {
	t.Helper()

	repoRoot = FindRepoRoot(t)
	storageRoot = filepath.Join(t.TempDir(), ".sticky")
	if _, err := RunMochiSticky(t, repoRoot, storageRoot, "init"); err != nil {
		t.Fatalf("init storage: %v", err)
	}
	return repoRoot, storageRoot
}

// RunMochiSticky executes `go run . --storage <root>` with the provided args.
func RunMochiSticky(t *testing.T, repoRoot, storageRoot string, args ...string) (string, error) {
	t.Helper()

	commandArgs := append([]string{"run", ".", "--storage", storageRoot}, args...)
	cmd := exec.Command("go", commandArgs...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "MOCHI_STICKY_STORAGE=")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return out.String(), formatCommandError(commandArgs, out.String(), err)
	}
	return out.String(), nil
}

// RunMochiStickyWithInput executes the CLI while providing stdin content.
func RunMochiStickyWithInput(
	t *testing.T,
	repoRoot,
	storageRoot,
	input string,
	args ...string,
) (string, error) {
	t.Helper()

	commandArgs := append([]string{"run", ".", "--storage", storageRoot}, args...)
	cmd := exec.Command("go", commandArgs...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "MOCHI_STICKY_STORAGE=")
	cmd.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return out.String(), formatCommandError(commandArgs, out.String(), err)
	}
	return out.String(), nil
}

func formatCommandError(args []string, output string, err error) error {
	if output == "" {
		return err
	}
	return fmt.Errorf("%w\ncommand: go %s\noutput:\n%s", err, strings.Join(args, " "), output)
}

// EditorCommandForTests returns a no-op editor command compatible with the OS.
func EditorCommandForTests() string {
	if runtime.GOOS == "windows" {
		return "cmd /c exit 0"
	}
	return "true"
}
