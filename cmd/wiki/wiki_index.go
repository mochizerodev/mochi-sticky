package wiki

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Generate wiki index from pages",
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

		includeTemplates, err := cmd.Flags().GetBool("include-templates")
		if err != nil {
			return err
		}
		templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
		if err != nil {
			return err
		}
		writeIndex, err := cmd.Flags().GetBool("write")
		if err != nil {
			return err
		}
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
		if output == "" {
			output = filepath.Join(root, "_index.yaml")
		}

		pages, err := wiki.ListPagesWithTemplatesRoot(root, includeTemplates, templatePaths.Wiki)
		if err != nil {
			return err
		}
		index, err := wiki.GenerateIndex(pages)
		if err != nil {
			return err
		}

		if writeIndex {
			if err := wiki.SaveIndex(output, index); err != nil {
				return err
			}
		}

		if !writeIndex {
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetEscapeHTML(false)
			return encoder.Encode(index)
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Generated index at %s\n", output)
		return err
	},
}

func init() {
	wikiCmd.AddCommand(wikiIndexCmd)
	wikiIndexCmd.Flags().Bool("include-templates", false, "Include template pages in index generation")
	wikiIndexCmd.Flags().Bool("write", true, "Write _index.yaml (set false to print JSON)")
	wikiIndexCmd.Flags().String("output", "", "Output path (defaults to <storage>/wiki/_index.yaml)")
}
