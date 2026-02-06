package wiki

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

func buildRootManifest(root, prefix string, index wiki.Index, indexErr error) (wiki.RootManifest, []wiki.Page, error) {
	if indexErr == nil {
		pages, err := wiki.ListPagesFromIndex(root, index)
		if err != nil {
			return wiki.RootManifest{}, nil, err
		}
		manifest, err := wiki.BuildManifest(index, pages)
		if err != nil {
			return wiki.RootManifest{}, nil, err
		}
		return wiki.RootManifest{Root: root, Prefix: prefix, Pages: manifest}, pages, nil
	}
	pages, err := wiki.ListPages(root)
	if err != nil {
		return wiki.RootManifest{}, nil, err
	}
	manifest := wiki.BuildManifestFromPages(pages)
	return wiki.RootManifest{Root: root, Prefix: prefix, Pages: manifest}, pages, nil
}

var wikiExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export wiki content",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return err
		}
		if strings.TrimSpace(format) == "" {
			format = "md"
		}
		if format != "md" && format != "pdf" {
			return fmt.Errorf("unsupported format: %s", format)
		}

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

		externalRoots, err := cmd.Flags().GetStringSlice("root")
		if err != nil {
			return err
		}
		pageSlug, err := cmd.Flags().GetString("page")
		if err != nil {
			return err
		}
		section, err := cmd.Flags().GetString("section")
		if err != nil {
			return err
		}
		pageSlug = strings.TrimSpace(pageSlug)
		section = strings.TrimSpace(section)
		if pageSlug != "" && section != "" {
			return fmt.Errorf("choose either --page or --section")
		}
		if (pageSlug != "" || section != "") && len(externalRoots) > 0 {
			return fmt.Errorf("--root cannot be used with --page or --section")
		}
		filterTitle, err := cmd.Flags().GetString("filter-title")
		if err != nil {
			return err
		}
		filterTagsRaw, err := cmd.Flags().GetString("filter-tags")
		if err != nil {
			return err
		}
		filterSection, err := cmd.Flags().GetString("filter-section")
		if err != nil {
			return err
		}
		filterQuery, err := cmd.Flags().GetString("filter-query")
		if err != nil {
			return err
		}
		filterTagMode, err := cmd.Flags().GetString("filter-tag-mode")
		if err != nil {
			return err
		}
		includeLinked, err := cmd.Flags().GetBool("include-linked")
		if err != nil {
			return err
		}
		linkTypes, err := cmd.Flags().GetStringSlice("link-type")
		if err != nil {
			return err
		}
		filterTags := splitTagFilter(filterTagsRaw)
		filtersActive := strings.TrimSpace(filterTitle) != "" ||
			strings.TrimSpace(filterSection) != "" ||
			strings.TrimSpace(filterQuery) != "" ||
			len(filterTags) > 0
		if filtersActive && len(externalRoots) > 0 {
			return fmt.Errorf("--filter-* flags cannot be used with --root")
		}

		rootManifests := make([]wiki.RootManifest, 0)
		mainPrefix, err := cmd.Flags().GetString("prefix")
		if err != nil {
			return err
		}

		mainManifest, pages, err := buildRootManifest(root, mainPrefix, index, indexErr)
		if err != nil {
			return err
		}
		selection := wiki.ExportSelection{Page: pageSlug, Section: section}
		if pageSlug != "" {
			manifest, err := wiki.BuildPageManifest(root, pages, pageSlug)
			if err != nil {
				return err
			}
			mainManifest.Pages = manifest
		}
		if section != "" {
			sectionIndex := index
			if indexErr != nil {
				sectionIndex, err = wiki.GenerateIndex(pages)
				if err != nil {
					return err
				}
			}
			if includeLinked {
				baseSection, _, err := wiki.FindSection(sectionIndex, section)
				if err != nil {
					return err
				}
				linked, err := wiki.ResolveLinkedSections(sectionIndex, baseSection, linkTypes)
				if err != nil {
					return err
				}
				sections := append([]wiki.IndexSection{baseSection}, linked...)
				manifest, err := wiki.BuildManifestForSections(sectionIndex, pages, sections)
				if err != nil {
					return err
				}
				mainManifest.Pages = manifest
			} else {
				manifest, err := wiki.BuildSectionManifest(sectionIndex, pages, section)
				if err != nil {
					return err
				}
				mainManifest.Pages = manifest
			}
		}

		if filtersActive {
			pageMap := make(map[string]wiki.Page, len(pages))
			for _, page := range pages {
				slug := strings.TrimSpace(page.Slug)
				if slug == "" && page.FilePath != "" {
					slug = slugFromPath(root, page.FilePath)
				}
				if slug == "" {
					continue
				}
				page.Slug = slug
				pageMap[slug] = page
			}
			mainManifest.Pages = wiki.FilterManifest(mainManifest.Pages, pageMap, wiki.FilterOptions{
				Title:           filterTitle,
				Tags:            filterTags,
				TagMode:         filterTagMode,
				Section:         filterSection,
				Query:           filterQuery,
				CaseInsensitive: true,
			})
		}

		rootManifests = append(rootManifests, mainManifest)
		if pageSlug == "" && section == "" {
			for _, external := range externalRoots {
				parts := strings.SplitN(external, ":", 2)
				pathPart := strings.TrimSpace(parts[0])
				if pathPart == "" {
					continue
				}
				if !filepath.IsAbs(pathPart) {
					pathPart = filepath.Join(workingDir, pathPart)
				}
				prefix := ""
				if len(parts) == 2 {
					prefix = strings.TrimSpace(parts[1])
				}
				externalManifest, err := wiki.BuildRootManifest(pathPart, prefix)
				if err != nil {
					return err
				}
				rootManifests = append(rootManifests, externalManifest)
			}

			if err := wiki.ValidateManifests(rootManifests); err != nil {
				return err
			}
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
		if strings.TrimSpace(output) == "" {
			output = wiki.DefaultExportPath(root, format, selection)
		}

		var data []byte
		if len(rootManifests) > 1 {
			manifest := wiki.FlattenManifests(rootManifests)
			if len(manifest) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No pages to export.")
				return err
			}
			data, err = wiki.ExportMarkdownMulti(rootManifests)
		} else {
			if len(mainManifest.Pages) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No pages to export.")
				return err
			}
			data, err = wiki.ExportMarkdown(root, mainManifest.Pages)
		}
		if err != nil {
			return err
		}

		if format == "md" {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			if err := wiki.WriteExportContext(ctx, output, data); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Exported wiki to %s\n", output)
			return err
		}

		title, err := cmd.Flags().GetString("title")
		if err != nil {
			return err
		}
		author, err := cmd.Flags().GetString("author")
		if err != nil {
			return err
		}
		template, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}
		if strings.TrimSpace(template) == "" {
			templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
			if err != nil {
				return err
			}
			template = templatePaths.WikiPDF
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if _, err := wiki.WritePDFContext(ctx, data, wiki.PDFOptions{
			Output:   output,
			Title:    title,
			Author:   author,
			Template: template,
			BaseDir:  workingDir,
			TempDir:  root,
		}); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Exported wiki to %s\n", output)
		return err
	},
}

func init() {
	wikiCmd.AddCommand(wikiExportCmd)
	wikiExportCmd.Flags().String("format", "md", "Export format (md|pdf)")
	wikiExportCmd.Flags().String("output", "", "Output path (defaults to <storage>/wiki/export.md)")
	wikiExportCmd.Flags().String("page", "", "Export a single page by slug")
	wikiExportCmd.Flags().String("section", "", "Export a section by slug or title")
	wikiExportCmd.Flags().String("filter-title", "", "Filter export by title substring")
	wikiExportCmd.Flags().String("filter-tags", "", "Filter export by tags (comma-separated)")
	wikiExportCmd.Flags().String("filter-tag-mode", "any", "Tag filter mode (any|all)")
	wikiExportCmd.Flags().String("filter-section", "", "Filter export by section")
	wikiExportCmd.Flags().String("filter-query", "", "Filter export by query")
	wikiExportCmd.Flags().Bool("include-linked", false, "Include linked sections when exporting a section")
	wikiExportCmd.Flags().StringSlice("link-type", nil, "Section link types to include (depends_on, related_to)")
	wikiExportCmd.Flags().String("title", "", "PDF title metadata")
	wikiExportCmd.Flags().String("author", "", "PDF author metadata")
	wikiExportCmd.Flags().String("template", "", "PDF template path (defaults to configured wiki_pdf template)")
	wikiExportCmd.Flags().StringSlice("root", nil, "Additional wiki roots (path[:prefix])")
	wikiExportCmd.Flags().String("prefix", "", "Prefix for main wiki root when combining")
}

func splitTagFilter(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	return tags
}
