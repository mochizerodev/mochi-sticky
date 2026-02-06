package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/wiki"

	"github.com/spf13/cobra"
)

var wikiCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a wiki page",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		root := wikiRoot(storageRoot)
		templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
		if err != nil {
			return err
		}

		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		slugInput, err := cmd.Flags().GetString("slug")
		if err != nil {
			return err
		}
		section, err := cmd.Flags().GetString("section")
		if err != nil {
			return err
		}
		order, err := cmd.Flags().GetInt("order")
		if err != nil {
			return err
		}
		tagsInput, err := cmd.Flags().GetString("tags")
		if err != nil {
			return err
		}
		status, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}

		slug := strings.TrimSpace(slugInput)
		if slug == "" {
			slug = wiki.Slugify(title)
		}
		if strings.TrimSpace(section) != "" && !strings.Contains(slug, "/") {
			slug = strings.TrimSuffix(strings.TrimSpace(section), "/") + "/" + slug
		}
		slug, err = wiki.NormalizeSlug(slug)
		if err != nil {
			return err
		}
		if strings.TrimSpace(status) == "" {
			status = "published"
		}

		var content string
		if strings.TrimSpace(templateName) != "" {
			templateDir := templatePaths.Wiki
			if strings.TrimSpace(templateDir) == "" {
				templateDir = filepath.Join(root, "templates")
			}
			templatePath := filepath.Join(templateDir, templateName+".md")
			templatePage, err := wiki.LoadPage(templatePath)
			if err != nil {
				return fmt.Errorf("template not found: %s", templateName)
			}
			if strings.TrimSpace(templatePage.Title) != "" {
				title = templatePage.Title
			}
			if strings.TrimSpace(templatePage.Content) != "" {
				content = templatePage.Content
			}
			if len(templatePage.Tags) > 0 && strings.TrimSpace(tagsInput) == "" {
				tagsInput = strings.Join(templatePage.Tags, ",")
			}
			if strings.TrimSpace(templatePage.Status) != "" && strings.TrimSpace(status) == "" {
				status = templatePage.Status
			}
		}

		page := wiki.Page{
			Title:   title,
			Slug:    slug,
			Section: strings.TrimSpace(section),
			Order:   order,
			Tags:    board.ParseTags(tagsInput),
			Status:  strings.TrimSpace(status),
			Content: content,
		}
		path, err := pagePath(root, slug)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("page already exists: %s", slug)
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := wiki.SavePage(path, page); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created wiki page %s\n", slug)
		return err
	},
}

func init() {
	wikiCmd.AddCommand(wikiCreateCmd)
	wikiCreateCmd.Flags().String("slug", "", "Explicit slug (defaults to slugified title)")
	wikiCreateCmd.Flags().String("section", "", "Optional section name")
	wikiCreateCmd.Flags().Int("order", 0, "Optional order within a section")
	wikiCreateCmd.Flags().String("tags", "", "Comma-separated tags")
	wikiCreateCmd.Flags().String("status", "published", "Status: draft|published|archived")
	wikiCreateCmd.Flags().String("template", "", "Template name (from configured wiki templates)")
}
