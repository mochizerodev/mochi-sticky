package board

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var boardAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new board",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
		if err != nil {
			return err
		}
		repo, err := cli.BoardRepoFromCwd()
		if err != nil {
			return err
		}
		name := strings.Join(args, " ")
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		createdBoard, err := repo.CreateBoardContext(ctx, name)
		if err != nil {
			return err
		}
		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}
		if strings.TrimSpace(templateName) != "" {
			templateDir := templatePaths.Board
			if strings.TrimSpace(templateDir) == "" {
				templateDir = filepath.Join(storageRoot, "templates", "board")
			}
			templatePath := filepath.Join(templateDir, templateName+".yaml")
			data, err := os.ReadFile(templatePath)
			if err != nil {
				return fmt.Errorf("template not found: %s", templateName)
			}
			cfg, err := board.ParseConfig(data)
			if err != nil {
				return err
			}
			cfg.NextID = 1
			boardRepo, err := board.NewRepositoryForBoardWithStorage(workingDir, createdBoard.ID, storageRoot)
			if err != nil {
				return err
			}
			if err := boardRepo.SaveConfigContext(ctx, cfg); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created board %s\n", createdBoard.ID)
		return err
	},
}

func init() {
	boardCmd.AddCommand(boardAddCmd)
	boardAddCmd.Flags().String("template", "", "Template name (from configured board templates)")
}
