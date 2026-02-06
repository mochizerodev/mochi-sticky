package wiki

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/shared"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List wiki pages",
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

		statusFilter, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}
		includeTemplates, err := cmd.Flags().GetBool("include-templates")
		if err != nil {
			return err
		}
		templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
		if err != nil {
			return err
		}
		titleFilter, err := cmd.Flags().GetString("title")
		if err != nil {
			return err
		}
		tagFilter, err := cmd.Flags().GetString("tags")
		if err != nil {
			return err
		}
		tagMode, err := cmd.Flags().GetString("tag-mode")
		if err != nil {
			return err
		}
		sectionFilter, err := cmd.Flags().GetString("section")
		if err != nil {
			return err
		}
		queryFilter, err := cmd.Flags().GetString("query")
		if err != nil {
			return err
		}
		filterTags := splitTagFilter(tagFilter)

		var pages []wiki.Page
		if indexErr == nil {
			pages, err = wiki.ListPagesFromIndex(root, index)
			if err != nil {
				return err
			}
			if includeTemplates {
				allPages, err := wiki.ListPagesWithTemplatesRoot(root, true, templatePaths.Wiki)
				if err != nil {
					return err
				}
				pagesBySlug := make(map[string]wiki.Page, len(allPages))
				for _, page := range allPages {
					slug := strings.TrimSpace(page.Slug)
					if slug == "" && page.FilePath != "" {
						if shared.IsSubpath(templatePaths.Wiki, page.FilePath) {
							slug = slugFromPath(templatePaths.Wiki, page.FilePath)
						} else {
							slug = slugFromPath(root, page.FilePath)
						}
					}
					if slug == "" {
						continue
					}
					page.Slug = slug
					pagesBySlug[slug] = page
				}
				ordered := make([]wiki.Page, 0, len(pagesBySlug))
				seen := make(map[string]struct{}, len(pagesBySlug))
				for _, page := range pages {
					slug := strings.TrimSpace(page.Slug)
					if slug == "" && page.FilePath != "" {
						if shared.IsSubpath(templatePaths.Wiki, page.FilePath) {
							slug = slugFromPath(templatePaths.Wiki, page.FilePath)
						} else {
							slug = slugFromPath(root, page.FilePath)
						}
					}
					if slug == "" {
						continue
					}
					if full, ok := pagesBySlug[slug]; ok {
						ordered = append(ordered, full)
					} else {
						ordered = append(ordered, page)
					}
					seen[slug] = struct{}{}
				}
				remaining := make([]wiki.Page, 0)
				for slug, page := range pagesBySlug {
					if _, ok := seen[slug]; ok {
						continue
					}
					remaining = append(remaining, page)
				}
				sort.Slice(remaining, func(i, j int) bool {
					return strings.ToLower(remaining[i].Title) < strings.ToLower(remaining[j].Title)
				})
				pages = append(ordered, remaining...)
			}
		} else {
			pages, err = wiki.ListPagesWithTemplatesRoot(root, includeTemplates, templatePaths.Wiki)
			if err != nil {
				return err
			}
			sort.Slice(pages, func(i, j int) bool {
				return strings.ToLower(pages[i].Title) < strings.ToLower(pages[j].Title)
			})
		}

		pages = wiki.FilterPagesByStatus(pages, statusFilter)
		pages = wiki.FilterPages(pages, wiki.FilterOptions{
			Title:           titleFilter,
			Tags:            filterTags,
			TagMode:         tagMode,
			Section:         sectionFilter,
			Query:           queryFilter,
			CaseInsensitive: true,
		})
		if len(pages) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No wiki pages found.")
			return err
		}

		for _, page := range pages {
			slug := strings.TrimSpace(page.Slug)
			if slug == "" && page.FilePath != "" {
				if shared.IsSubpath(templatePaths.Wiki, page.FilePath) {
					slug = slugFromPath(templatePaths.Wiki, page.FilePath)
				} else {
					slug = slugFromPath(root, page.FilePath)
				}
			}
			if slug == "" {
				slug = "(missing slug)"
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", slug, page.Title); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	wikiCmd.AddCommand(wikiListCmd)
	wikiListCmd.Flags().String("status", "", "Filter by status (draft|published|archived)")
	wikiListCmd.Flags().Bool("include-templates", false, "Include template pages in list output")
	wikiListCmd.Flags().String("title", "", "Filter by title substring")
	wikiListCmd.Flags().String("tags", "", "Filter by tags (comma-separated)")
	wikiListCmd.Flags().String("tag-mode", "any", "Tag filter mode (any|all)")
	wikiListCmd.Flags().String("section", "", "Filter by section")
	wikiListCmd.Flags().String("query", "", "Filter by keyword query")
}
