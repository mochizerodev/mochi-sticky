package board

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var boardRenameCmd = &cobra.Command{
	Use:   "rename <id> <name>",
	Short: "Rename a board",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := cli.BoardRepoFromCwd()
		if err != nil {
			return err
		}
		id := args[0]
		name := strings.Join(args[1:], " ")
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		updatedBoard, err := repo.RenameBoardContext(ctx, id, name)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Renamed board %s to %s\n", updatedBoard.ID, updatedBoard.Name)
		return err
	},
}

func init() {
	boardCmd.AddCommand(boardRenameCmd)
}
