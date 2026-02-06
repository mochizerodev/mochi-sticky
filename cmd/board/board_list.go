package board

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List boards",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := cli.BoardRepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		boards, active, err := repo.ListBoardsContext(ctx)
		if err != nil {
			return err
		}
		if len(boards) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No boards found.")
			return err
		}
		for _, boardItem := range boards {
			activeMark := " "
			if boardItem.ID == active {
				activeMark = "*"
			}
			archived := ""
			if boardItem.Archived {
				archived = " (archived)"
			}
			if _, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"%s %s - %s%s\n",
				activeMark,
				boardItem.ID,
				boardItem.Name,
				archived,
			); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	boardCmd.AddCommand(boardListCmd)
}
