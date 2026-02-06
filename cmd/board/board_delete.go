package board

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var boardDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cli.RequireConfirm(cmd, fmt.Sprintf("Delete board %q? This cannot be undone.", args[0])); err != nil {
			return err
		}
		repo, err := cli.BoardRepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.DeleteBoardContext(ctx, args[0]); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted board %s\n", args[0])
		return err
	},
}

func init() {
	boardDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	boardCmd.AddCommand(boardDeleteCmd)
}
