package wiki

import (
	"fmt"
	"os"
	"path/filepath"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a wiki page",
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
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("page not found: %s", slug)
			}
			return err
		}

		updateIndex, err := cmd.Flags().GetBool("update-index")
		if err != nil {
			return err
		}
		if updateIndex {
			indexPath := filepath.Join(root, "_index.yaml")
			index, err := wiki.LoadIndex(indexPath)
			if err == nil {
				normalized, err := wiki.NormalizeSlug(slug)
				if err != nil {
					return err
				}
				if index.RemoveSlug(normalized) {
					if err := wiki.SaveIndex(indexPath, index); err != nil {
						return err
					}
				}
			}
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted wiki page %s\n", slug)
		return err
	},
}

func init() {
	wikiCmd.AddCommand(wikiDeleteCmd)
	wikiDeleteCmd.Flags().Bool("update-index", false, "Remove page from _index.yaml when present")
}
