package adr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/storage"

	"github.com/spf13/cobra"
)

var adrEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an ADR in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := adr.ParseID(args[0])
		if err != nil {
			return err
		}

		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		root := adrRoot(storageRoot)
		repo, err := adr.NewRepository(root)
		if err != nil {
			return err
		}
		record, err := repo.GetADRByID(id)
		if err != nil {
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
		editCmd := exec.Command(parts[0], append(parts[1:], record.FilePath)...)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr
		return editCmd.Run()
	},
}

func init() {
	adrCmd.AddCommand(adrEditCmd)
	adrEditCmd.Flags().String("editor", "", "Override the editor command")
}
