package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	boardpkg "mochi-sticky/internal/board"
	"mochi-sticky/internal/storage"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var storageRootFlagRef *string

func SetStorageRootFlagRef(ref *string) {
	storageRootFlagRef = ref
}

func storageFlagValue() string {
	if storageRootFlagRef == nil {
		return ""
	}
	return *storageRootFlagRef
}

func ResolveStorageRoot(workingDir string, allowMissing bool) (string, error) {
	return storage.ResolveRoot(workingDir, allowMissing, storageFlagValue())
}

func LoadStorageConfig(workingDir string) (storage.Config, error) {
	return storage.LoadConfigWithOverride(workingDir, storageFlagValue())
}

func ResolveTemplatePaths(workingDir, storageRoot string) (storage.TemplatePaths, error) {
	cfg, err := storage.LoadConfigWithOverride(workingDir, storageFlagValue())
	if err != nil {
		return storage.TemplatePaths{}, err
	}
	return storage.ResolveTemplates(workingDir, storageRoot, cfg)
}

func RepoFromCwd() (*boardpkg.Repository, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	storageRoot, err := ResolveStorageRoot(workingDir, false)
	if err != nil {
		return nil, err
	}
	return boardpkg.NewRepositoryWithStorage(workingDir, storageRoot)
}

func BoardRepoFromCwd() (*boardpkg.BoardRepository, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	storageRoot, err := ResolveStorageRoot(workingDir, false)
	if err != nil {
		return nil, err
	}
	return boardpkg.NewBoardRepositoryWithStorage(workingDir, storageRoot)
}

func PrintBoardContext(out io.Writer, ctx boardpkg.BoardContext) error {
	if ctx.Scope == "" && ctx.Release == "" && ctx.Target == "" && ctx.Notes == "" && len(ctx.Owners) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, "Context:"); err != nil {
		return err
	}
	addLine := func(label, value string) error {
		if strings.TrimSpace(value) == "" {
			return nil
		}
		_, err := fmt.Fprintf(out, "  %s: %s\n", label, value)
		return err
	}
	if err := addLine("Scope", ctx.Scope); err != nil {
		return err
	}
	if err := addLine("Release", ctx.Release); err != nil {
		return err
	}
	if err := addLine("Target", ctx.Target); err != nil {
		return err
	}
	if len(ctx.Owners) > 0 {
		if _, err := fmt.Fprintf(out, "  Owners: %s\n", strings.Join(ctx.Owners, ", ")); err != nil {
			return err
		}
	}
	if err := addLine("Notes", ctx.Notes); err != nil {
		return err
	}
	return nil
}

func ConfirmPrompt(cmd *cobra.Command, message string) (bool, error) {
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N]: ", message); err != nil {
		return false, err
	}
	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(response))
	return answer == "y" || answer == "yes", nil
}

func RequireConfirm(cmd *cobra.Command, message string) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}
	if force {
		return nil
	}
	ok, err := ConfirmPrompt(cmd, message)
	if err != nil {
		return err
	}
	if !ok {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Canceled."); err != nil {
			return err
		}
		return fmt.Errorf("confirmation canceled")
	}
	return nil
}

func WikiRoot(storageRoot string) string {
	return filepath.Join(storageRoot, "wiki")
}

// AdrRoot returns the ADR root directory within the storage root.
func AdrRoot(storageRoot string) string {
	return filepath.Join(storageRoot, "adrs")
}

func PagePath(root, slug string) (string, error) {
	clean, err := wiki.NormalizeSlug(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, filepath.FromSlash(clean)+".md"), nil
}

func SlugFromPath(root, pagePath string) string {
	return wiki.SlugFromPath(root, pagePath)
}

func ResolvePathForConfig(workingDir, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if !filepath.IsAbs(trimmed) {
		trimmed = filepath.Join(workingDir, trimmed)
	}
	path, err := filepath.Abs(trimmed)
	if err != nil {
		return trimmed
	}
	return path
}

func EnsureReadableDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("missing directory: %s", path)
		}
		return fmt.Errorf("unable to access directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("expected directory but found file: %s", path)
	}
	return nil
}

func EnsureReadableFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("missing file: %s", path)
		}
		return fmt.Errorf("unable to access file %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected file but found directory: %s", path)
	}
	return nil
}

func ResolveEditorOverride(cliEditor, configEditor string) string {
	trimmed := strings.TrimSpace(cliEditor)
	if trimmed != "" {
		return trimmed
	}
	if env := strings.TrimSpace(os.Getenv("MOCHI_EDITOR")); env != "" {
		return env
	}
	if env := strings.TrimSpace(os.Getenv("EDITOR")); env != "" {
		return env
	}
	return strings.TrimSpace(configEditor)
}
