package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveEditorOverride(t *testing.T) {
	cases := []struct {
		name         string
		mochiEditor  string
		editor       string
		cliOverride  string
		configEditor string
		expect       string
	}{
		{
			name:         "cli override",
			mochiEditor:  "",
			editor:       "",
			cliOverride:  "vim",
			configEditor: "nano",
			expect:       "vim",
		},
		{
			name:         "mochi editor env",
			mochiEditor:  "mochi",
			editor:       "",
			cliOverride:  "",
			configEditor: "nano",
			expect:       "mochi",
		},
		{
			name:         "editor env",
			mochiEditor:  "",
			editor:       "ed",
			cliOverride:  "",
			configEditor: "nano",
			expect:       "ed",
		},
		{
			name:         "config editor",
			mochiEditor:  "",
			editor:       "",
			cliOverride:  "",
			configEditor: "nano",
			expect:       "nano",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			t.Setenv("MOCHI_EDITOR", tc.mochiEditor)
			t.Setenv("EDITOR", tc.editor)

			// Act
			got := ResolveEditorOverride(tc.cliOverride, tc.configEditor)

			// Assert
			if got != tc.expect {
				t.Fatalf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestResolvePathForConfig(t *testing.T) {
	// Arrange
	baseDir := t.TempDir()
	relative := filepath.Join("configs", "thing")
	expected := filepath.Join(baseDir, relative)
	abs := filepath.Join(baseDir, "abs")

	// Act
	relativePath := ResolvePathForConfig(baseDir, relative)
	emptyPath := ResolvePathForConfig(baseDir, "")
	absolutePath := ResolvePathForConfig(baseDir, abs)

	// Assert
	if relativePath != expected {
		t.Fatalf("expected %q, got %q", expected, relativePath)
	}
	if emptyPath != "" {
		t.Fatalf("expected empty value to return empty path")
	}
	if absolutePath != abs {
		t.Fatalf("expected absolute path to be unchanged")
	}
}

func TestEnsureReadableDirAndFile(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Act
	dirErr := EnsureReadableDir(dir)
	fileErr := EnsureReadableFile(file)
	fileAsDirErr := EnsureReadableDir(file)
	dirAsFileErr := EnsureReadableFile(dir)

	// Assert
	if dirErr != nil {
		t.Fatalf("expected dir to be readable: %v", dirErr)
	}
	if fileErr != nil {
		t.Fatalf("expected file to be readable: %v", fileErr)
	}
	if fileAsDirErr == nil {
		t.Fatalf("expected file to be rejected as dir")
	}
	if dirAsFileErr == nil {
		t.Fatalf("expected dir to be rejected as file")
	}
}

func TestConfirmPrompt(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect bool
	}{
		{name: "affirmative", input: "y\n", expect: true},
		{name: "negative", input: "no\n", expect: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			cmd := &cobra.Command{}
			cmd.SetIn(bytes.NewBufferString(tc.input))
			cmd.SetOut(&bytes.Buffer{})

			// Act
			ok, err := ConfirmPrompt(cmd, "Continue?")

			// Assert
			if err != nil {
				t.Fatalf("confirm prompt error: %v", err)
			}
			if ok != tc.expect {
				t.Fatalf("expected confirmation %v, got %v", tc.expect, ok)
			}
		})
	}
}

func TestRequireConfirm(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		forceValue  string
		expectError bool
	}{
		{name: "cancelled", input: "n\n", forceValue: "false", expectError: true},
		{name: "forced", input: "", forceValue: "true", expectError: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			cmd := &cobra.Command{}
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.input))
			cmd.Flags().Bool("force", false, "force")
			if err := cmd.Flags().Set("force", tc.forceValue); err != nil {
				t.Fatalf("set force: %v", err)
			}

			// Act
			err := RequireConfirm(cmd, "Proceed?")

			// Assert
			if tc.expectError && err == nil {
				t.Fatalf("expected confirmation canceled error")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected force to skip prompt, got %v", err)
			}
		})
	}
}

func TestResolveStorageRootUsesOverrideFlag(t *testing.T) {
	// Arrange
	baseDir := t.TempDir()
	overrideDir := filepath.Join(baseDir, "override")
	if err := os.MkdirAll(overrideDir, 0o755); err != nil {
		t.Fatalf("mkdir override: %v", err)
	}
	flagValue := "override"
	SetStorageRootFlagRef(&flagValue)
	t.Cleanup(func() { SetStorageRootFlagRef(nil) })

	// Act
	resolved, err := ResolveStorageRoot(baseDir, false)

	// Assert
	if err != nil {
		t.Fatalf("resolve storage root: %v", err)
	}
	if !strings.HasSuffix(resolved, "override") {
		t.Fatalf("expected override path, got %q", resolved)
	}
}
