package board

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	boardpkg "mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var boardShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show board details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}

		repo, err := boardpkg.NewBoardRepositoryWithStorage(workingDir, storageRoot)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return err
		}
		id := args[0]
		for _, boardItem := range boards {
			if boardItem.ID != id {
				continue
			}
			boardRepo, err := boardpkg.NewRepositoryForBoardWithStorage(workingDir, boardItem.ID, storageRoot)
			if err != nil {
				return err
			}
			description, err := boardRepo.LoadBoardDescriptionContext(ctx)
			if err != nil {
				return err
			}
			activeMark := ""
			if boardItem.ID == active {
				activeMark = " (active)"
			}
			archived := "false"
			if boardItem.Archived {
				archived = "true"
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\nName: %s\n", boardItem.ID, boardItem.Name); err != nil {
				return err
			}
			if strings.TrimSpace(description) != "" {
				if _, err := fmt.Fprintf(
					cmd.OutOrStdout(),
					"Description:\n%s\n",
					strings.TrimRight(description, "\n"),
				); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Description: (empty)"); err != nil {
					return err
				}
			}
			config, err := boardRepo.LoadConfigContext(ctx)
			if err != nil {
				return err
			}
			if err := cli.PrintBoardContext(cmd.OutOrStdout(), config.Context); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"Path: %s\nArchived: %s%s\nCreated: %s\n",
				boardItem.Path,
				archived,
				activeMark,
				boardItem.Created.Format("2006-01-02"),
			); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("board not found: %s", id)
	},
}

func init() {
	boardCmd.AddCommand(boardShowCmd)
}
