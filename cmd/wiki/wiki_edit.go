package wiki

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/storage"

	"github.com/spf13/cobra"
)

var wikiEditCmd = &cobra.Command{
	Use:   "edit <slug>",
	Short: "Edit a wiki page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		root := wikiRoot(storageRoot)
		path, err := pagePath(root, slug)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("page not found: %s", slug)
			}
			return err
		}

		cfg, err := storage.LoadConfigFromRoot(storageRoot)
		if err != nil {
			return err
		}
		editorOverride, err := cmd.Flags().GetString("editor")
		if err != nil {
			return err
		}
		editor := cli.ResolveEditorOverride(editorOverride, cfg.Editor)
		if strings.TrimSpace(editor) == "" {
			editor = "nano"
		}

		parts := strings.Fields(editor)
		if len(parts) == 0 {
			return fmt.Errorf("editor is required")
		}

		editCmd := exec.Command(parts[0], append(parts[1:], path)...)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr
		return editCmd.Run()
	},
}

func init() {
	wikiCmd.AddCommand(wikiEditCmd)
	wikiEditCmd.Flags().String("editor", "", "Override the editor command")
}
