package wiki

import (
	"fmt"
	"os"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiViewCmd = &cobra.Command{
	Use:   "view <slug>",
	Short: "View a wiki page",
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
		page, err := wiki.LoadPage(path)
		if err != nil {
			return err
		}
		if page.Content == "" {
			return nil
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), page.Content)
		return err
	},
}

func init() {
	wikiCmd.AddCommand(wikiViewCmd)
}
