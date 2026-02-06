package wiki

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiSectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "List wiki sections",
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

		index, err := wiki.LoadIndex(indexPath)
		if err != nil {
			if errors.Is(err, wiki.ErrIndexNotFound) {
				pages, err := wiki.ListPages(root)
				if err != nil {
					return err
				}
				index, err = wiki.GenerateIndex(pages)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		tagFilter, err := cmd.Flags().GetString("tags")
		if err != nil {
			return err
		}
		tagMode, err := cmd.Flags().GetString("tag-mode")
		if err != nil {
			return err
		}
		linkType, err := cmd.Flags().GetString("link-type")
		if err != nil {
			return err
		}
		linkTarget, err := cmd.Flags().GetString("link-target")
		if err != nil {
			return err
		}
		filterTags := splitTagFilter(tagFilter)

		sections := wiki.FilterSections(index.Sections, wiki.SectionFilterOptions{
			Tags:       filterTags,
			TagMode:    tagMode,
			LinkType:   linkType,
			LinkTarget: linkTarget,
		})
		if len(sections) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No wiki sections found.")
			return err
		}

		for _, section := range sections {
			slug := strings.TrimSpace(section.Slug)
			if slug == "" {
				slug = "(root)"
			}
			parts := []string{slug, section.Title}
			if len(section.Tags) > 0 {
				parts = append(parts, "tags:"+strings.Join(section.Tags, ", "))
			}
			if len(section.Links.DependsOn) > 0 {
				parts = append(parts, "depends_on:"+strings.Join(section.Links.DependsOn, ", "))
			}
			if len(section.Links.RelatedTo) > 0 {
				parts = append(parts, "related_to:"+strings.Join(section.Links.RelatedTo, ", "))
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), strings.Join(parts, "\t")); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	wikiCmd.AddCommand(wikiSectionsCmd)
	wikiSectionsCmd.Flags().String("tags", "", "Filter by tags (comma-separated)")
	wikiSectionsCmd.Flags().String("tag-mode", "any", "Tag filter mode (any|all)")
	wikiSectionsCmd.Flags().String("link-type", "", "Filter by link type (depends_on|related_to)")
	wikiSectionsCmd.Flags().String("link-target", "", "Filter by link target")
}
