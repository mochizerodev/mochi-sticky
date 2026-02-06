package tui

import (
	"os"
	"strings"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/storage"
	"mochi-sticky/internal/tui"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive board",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		editorOverride, err := cmd.Flags().GetString("editor")
		if err != nil {
			return err
		}
		setEditor, err := cmd.Flags().GetString("set-editor")
		if err != nil {
			return err
		}
		cfg, err := storage.LoadConfigFromRoot(storageRoot)
		if err != nil {
			return err
		}
		if strings.TrimSpace(setEditor) != "" {
			cfg.Editor = strings.TrimSpace(setEditor)
			if err := storage.SaveConfigToRoot(storageRoot, cfg); err != nil {
				return err
			}
		}

		editor := cli.ResolveEditorOverride(editorOverride, cfg.Editor)
		repo, err := board.NewRepositoryWithStorage(workingDir, storageRoot)
		if err != nil {
			return err
		}
		return tui.RunWithEditor(repo, editor)
	},
}

func Register(root *cobra.Command) {
	root.AddCommand(tuiCmd)
	tuiCmd.Flags().String("editor", "", "Override the editor command")
	tuiCmd.Flags().String("set-editor", "", "Persist the editor command in settings")
}
