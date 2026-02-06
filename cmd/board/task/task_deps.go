package taskcmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var (
	depsSetFlag string
)

var taskDepsCmd = &cobra.Command{
	Use:   "deps <id>",
	Short: "Show or set task dependencies",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
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

		if strings.TrimSpace(depsSetFlag) != "" {
			ids := board.ParseTags(depsSetFlag) // reuse tag parsing (comma or spaces)
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			if err := repo.UpdateTaskDependenciesContext(ctx, id, ids); err != nil {
				return err
			}
		}

		task, err := repo.GetTaskByID(id)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Task: %s\nDepends on: %s\n", id, strings.Join(task.DependsOn, ", "))
		return err
	},
}

func init() {
	taskDepsCmd.Flags().StringVar(&depsSetFlag, "set", "", "Comma-separated list of dependency IDs to set")
	taskCmd.AddCommand(taskDepsCmd)
}
