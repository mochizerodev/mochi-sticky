package adr

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var adrCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create an ADR",
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
		root := adrRoot(storageRoot)
		templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
		if err != nil {
			return err
		}
		repo, err := adr.NewRepository(root)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.InitStoreContext(ctx); err != nil {
			return err
		}

		status, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}
		dateRaw, err := cmd.Flags().GetString("date")
		if err != nil {
			return err
		}
		when, err := parseDate(dateRaw)
		if err != nil {
			return err
		}
		tagsRaw, err := cmd.Flags().GetString("tags")
		if err != nil {
			return err
		}
		linksRaw, err := cmd.Flags().GetString("links")
		if err != nil {
			return err
		}
		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}
		bodyRaw, err := cmd.Flags().GetString("body")
		if err != nil {
			return err
		}

		var body string
		if strings.TrimSpace(bodyRaw) == "-" {
			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}
			body = string(data)
		} else if strings.TrimSpace(bodyRaw) != "" {
			body = bodyRaw
		} else if strings.TrimSpace(templateName) != "" {
			templateDir := templatePaths.ADR
			if strings.TrimSpace(templateDir) == "" {
				templateDir = filepath.Join(root, adr.TemplatesDirName)
			}
			templatePath := filepath.Join(templateDir, templateName+".md")
			data, err := os.ReadFile(templatePath)
			if err != nil {
				return fmt.Errorf("template not found: %s", templateName)
			}
			body = string(data)
		}
		if strings.TrimSpace(body) == "" {
			body = adr.DefaultContent()
		}
		if err := adr.ValidateRequiredHeadings(body); err != nil {
			return err
		}

		record, err := repo.CreateADRContext(ctx, title, adr.CreateOptions{
			Status: status,
			Date:   when,
			Tags:   board.ParseTags(tagsRaw),
			Links:  splitCommaList(linksRaw),
			Body:   body,
		})
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created ADR %s\n", adr.FormatID(record.ID))
		return err
	},
}

func init() {
	adrCmd.AddCommand(adrCreateCmd)
	adrCreateCmd.Flags().String("status", "", "ADR status (defaults to first configured column)")
	adrCreateCmd.Flags().String("date", "", "Decision date (YYYY-MM-DD, defaults to today)")
	adrCreateCmd.Flags().String("tags", "", "Comma-separated tags")
	adrCreateCmd.Flags().String("links", "", "Comma-separated links (URLs, task IDs, wiki slugs, etc.)")
	adrCreateCmd.Flags().String("template", "", "Template name (from configured ADR templates)")
	adrCreateCmd.Flags().String("body", "", "ADR body markdown (use '-' to read from stdin)")
}
