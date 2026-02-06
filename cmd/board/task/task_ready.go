package taskcmd

import (
	"fmt"
	"os"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var taskReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "List tasks whose dependencies are satisfied",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		tasks, err := repo.ListReadyTasks()
		if err != nil {
			return err
		}
		for _, t := range tasks {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", t.ID, t.Title); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	taskCmd.AddCommand(taskReadyCmd)
}
