package adr

import (
	"fmt"
	"os"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var adrViewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "View an ADR",
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
		data, err := os.ReadFile(record.FilePath)
		if err != nil {
			return fmt.Errorf("failed to read adr %s: %w", record.FilePath, err)
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(data))
		return err
	},
}

func init() {
	adrCmd.AddCommand(adrViewCmd)
}
