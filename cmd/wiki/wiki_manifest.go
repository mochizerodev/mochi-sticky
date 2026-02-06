package wiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiManifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Generate an ordered wiki manifest",
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
		indexPath := filepath.Join(root, "_index.yaml")

		index, indexErr := wiki.LoadIndex(indexPath)
		if indexErr != nil && !errors.Is(indexErr, wiki.ErrIndexNotFound) {
			return indexErr
		}

		var pages []wiki.Page
		if indexErr == nil {
			pages, indexErr = wiki.ListPagesFromIndex(root, index)
			if indexErr != nil {
				return indexErr
			}
		} else {
			pages, indexErr = wiki.ListPages(root)
			if indexErr != nil {
				return indexErr
			}
		}

		var manifest []wiki.ManifestEntry
		if indexErr == nil {
			manifest, indexErr = wiki.BuildManifest(index, pages)
			if indexErr != nil {
				return indexErr
			}
		} else {
			manifest = wiki.BuildManifestFromPages(pages)
		}

		if len(manifest) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No pages to export.")
			return err
		}

		sort.SliceStable(manifest, func(i, j int) bool {
			if manifest[i].Order != manifest[j].Order {
				return manifest[i].Order < manifest[j].Order
			}
			return strings.ToLower(manifest[i].Title) < strings.ToLower(manifest[j].Title)
		})

		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetEscapeHTML(false)
		return encoder.Encode(manifest)
	},
}

func init() {
	wikiCmd.AddCommand(wikiManifestCmd)
}
