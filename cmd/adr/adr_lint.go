package adr

import (
	"fmt"
	"os"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var adrLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate ADR frontmatter and required sections",
	RunE: func(cmd *cobra.Command, args []string) error {
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
		adrs, err := repo.ListADRs()
		if err != nil {
			return err
		}
		if len(adrs) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No ADRs found.")
			return err
		}

		failures := 0
		for _, record := range adrs {
			if err := adr.ValidateADR(record); err != nil {
				failures++
				if _, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "%s: %v\n", adr.FormatID(record.ID), err); writeErr != nil {
					return writeErr
				}
			}
		}
		if failures > 0 {
			return fmt.Errorf("adr lint failed: %d issue(s)", failures)
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "OK")
		return err
	},
}

func init() {
	adrCmd.AddCommand(adrLintCmd)
}
