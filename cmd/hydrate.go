package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/storage"

	"github.com/spf13/cobra"
)

type hydrateReport struct {
	StorageRoot    string            `json:"storage_root"`
	ActiveBoard    string            `json:"active_board,omitempty"`
	BoardCount     int               `json:"board_count"`
	WikiRoot       string            `json:"wiki_root,omitempty"`
	ADRRoot        string            `json:"adr_root,omitempty"`
	PDFTemplate    string            `json:"pdf_template,omitempty"`
	TemplatesRoot  string            `json:"templates_root,omitempty"`
	TemplateDirs   map[string]string `json:"template_dirs,omitempty"`
	TemplateFiles  map[string]string `json:"template_files,omitempty"`
	Warnings       []string          `json:"warnings,omitempty"`
	ValidatedPaths []string          `json:"validated_paths,omitempty"`
}

var hydrateCmd = &cobra.Command{
	Use:   "hydrate",
	Short: "Validate and summarize storage configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := resolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}

		report := hydrateReport{
			StorageRoot: storageRoot,
			WikiRoot:    cli.WikiRoot(storageRoot),
			ADRRoot:     cli.AdrRoot(storageRoot),
		}

		config, err := loadStorageConfig(workingDir)
		if err != nil {
			return err
		}
		if strings.TrimSpace(config.PDFTemplate) != "" {
			report.PDFTemplate = cli.ResolvePathForConfig(workingDir, config.PDFTemplate)
		}
		templatePaths, err := storage.ResolveTemplates(workingDir, storageRoot, config)
		if err != nil {
			return err
		}
		if report.PDFTemplate == "" {
			report.PDFTemplate = templatePaths.WikiPDF
		}
		report.TemplatesRoot = templatePaths.Root
		report.TemplateDirs = map[string]string{
			"adr":   templatePaths.ADR,
			"task":  templatePaths.Task,
			"board": templatePaths.Board,
			"wiki":  templatePaths.Wiki,
		}
		report.TemplateFiles = map[string]string{
			"wiki_pdf": templatePaths.WikiPDF,
		}

		boardRepo, err := board.NewBoardRepositoryWithStorage(workingDir, storageRoot)
		if err != nil {
			return err
		}
		boards, active, err := boardRepo.ListBoards()
		if err != nil {
			if errors.Is(err, board.ErrStoreNotInitialized) {
				return fmt.Errorf("hydrate: storage not initialized at %s (run `mochi-sticky init`)", storageRoot)
			}
			return err
		}

		report.ActiveBoard = active
		report.BoardCount = len(boards)
		for _, board := range boards {
			if strings.TrimSpace(board.ID) == "" {
				report.Warnings = append(report.Warnings, "board registry contains an empty board id")
				continue
			}
			boardDir := filepath.Join(storageRoot, board.Path)
			if err := cli.EnsureReadableDir(boardDir); err != nil {
				report.Warnings = append(report.Warnings, err.Error())
				continue
			}
			report.ValidatedPaths = append(report.ValidatedPaths, boardDir)
			configPath := filepath.Join(boardDir, "config.yaml")
			if err := cli.EnsureReadableFile(configPath); err != nil {
				report.Warnings = append(report.Warnings, err.Error())
			} else {
				report.ValidatedPaths = append(report.ValidatedPaths, configPath)
			}
			if board.ID == active {
				tasksDir := filepath.Join(boardDir, "tasks")
				if err := cli.EnsureReadableDir(tasksDir); err != nil {
					report.Warnings = append(report.Warnings, err.Error())
				} else {
					report.ValidatedPaths = append(report.ValidatedPaths, tasksDir)
				}
			}
		}

		if err := cli.EnsureReadableDir(report.WikiRoot); err != nil {
			report.Warnings = append(report.Warnings, err.Error())
		} else {
			report.ValidatedPaths = append(report.ValidatedPaths, report.WikiRoot)
		}
		if err := cli.EnsureReadableDir(report.ADRRoot); err != nil {
			report.Warnings = append(report.Warnings, err.Error())
		} else {
			report.ValidatedPaths = append(report.ValidatedPaths, report.ADRRoot)
			configPath := filepath.Join(report.ADRRoot, "config.yaml")
			if err := cli.EnsureReadableFile(configPath); err != nil {
				report.Warnings = append(report.Warnings, err.Error())
			} else {
				report.ValidatedPaths = append(report.ValidatedPaths, configPath)
			}
		}
		if report.PDFTemplate != "" {
			if err := cli.EnsureReadableFile(report.PDFTemplate); err != nil {
				report.Warnings = append(report.Warnings, err.Error())
			} else {
				report.ValidatedPaths = append(report.ValidatedPaths, report.PDFTemplate)
			}
		}
		if report.TemplateDirs != nil {
			for _, path := range report.TemplateDirs {
				if strings.TrimSpace(path) == "" {
					continue
				}
				if err := cli.EnsureReadableDir(path); err != nil {
					report.Warnings = append(report.Warnings, err.Error())
				} else {
					report.ValidatedPaths = append(report.ValidatedPaths, path)
				}
			}
		}
		if report.TemplateFiles != nil {
			for _, path := range report.TemplateFiles {
				if strings.TrimSpace(path) == "" {
					continue
				}
				if err := cli.EnsureReadableFile(path); err != nil {
					report.Warnings = append(report.Warnings, err.Error())
				} else {
					report.ValidatedPaths = append(report.ValidatedPaths, path)
				}
			}
		}

		asJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}
		if asJSON {
			pretty, err := cmd.Flags().GetBool("pretty")
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetEscapeHTML(false)
			if pretty {
				encoder.SetIndent("", "  ")
			}
			return encoder.Encode(report)
		}

		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Storage: %s\n", report.StorageRoot); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Active board: %s\n", report.ActiveBoard); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Boards: %d\n", report.BoardCount); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Wiki root: %s\n", report.WikiRoot); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "ADR root: %s\n", report.ADRRoot); err != nil {
			return err
		}
		if report.PDFTemplate != "" {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "PDF template: %s\n", report.PDFTemplate); err != nil {
				return err
			}
		}
		if report.TemplatesRoot != "" {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Templates root: %s\n", report.TemplatesRoot); err != nil {
				return err
			}
		}
		if len(report.Warnings) > 0 {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Warnings:"); err != nil {
				return err
			}
			for _, warning := range report.Warnings {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), " - %s\n", warning); err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(hydrateCmd)
	hydrateCmd.Flags().Bool("json", false, "Output as JSON")
	hydrateCmd.Flags().Bool("pretty", false, "Pretty-print JSON output (requires --json)")
}
