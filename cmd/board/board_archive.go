package board

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var boardArchiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive a board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cli.RequireConfirm(cmd, fmt.Sprintf("Archive board %q?", args[0])); err != nil {
			return err
		}
		repo, err := cli.BoardRepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		archivedBoard, err := repo.ArchiveBoardContext(ctx, args[0])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Archived board %s\n", archivedBoard.ID)
		return err
	},
}

func init() {
	boardArchiveCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	boardCmd.AddCommand(boardArchiveCmd)
}
