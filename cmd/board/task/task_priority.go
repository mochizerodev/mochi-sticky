package taskcmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var priorityCmd = &cobra.Command{
	Use:   "priority <id> <priority>",
	Short: "Update task priority",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		priority, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid priority %q: %w", args[1], err)
		}
		if priority < 1 || priority > 3 {
			return fmt.Errorf("invalid priority %d (expected 1-3)", priority)
		}
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
		if err := repo.UpdateTaskPriorityContext(ctx, id, priority); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Updated priority for %s to %d\n", id, priority)
		return err
	},
}

func init() {
	taskCmd.AddCommand(priorityCmd)
}
