package board

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var boardUseCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Set the active board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := cli.BoardRepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.SetActiveBoardContext(ctx, args[0]); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Active board set to %s\n", args[0])
		return err
	},
}

func init() {
	boardCmd.AddCommand(boardUseCmd)
}
