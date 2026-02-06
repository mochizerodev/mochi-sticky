package wiki

import (
	"encoding/json"
	"fmt"
	"os"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint wiki pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		root := wikiRoot(storageRoot)
		pages, err := wiki.ListPages(root)
		if err != nil {
			return err
		}
		issues := wiki.LintPages(pages)
		if len(issues) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No issues found.")
			return err
		}
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetEscapeHTML(false)
		return encoder.Encode(issues)
	},
}

func init() {
	wikiCmd.AddCommand(wikiLintCmd)
}
