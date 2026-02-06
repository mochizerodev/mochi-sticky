package taskcmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <id> <status>",
	Short: "Move a task to a new status",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		status := args[1]

		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		repo, err := board.NewRepositoryWithStorage(workingDir, storageRoot)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.UpdateTaskStatusContext(ctx, id, status); err != nil {
			return err
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Moved task %s to %s\n", id, status)
		return err
	},
}

func init() {
	taskCmd.AddCommand(moveCmd)
}
