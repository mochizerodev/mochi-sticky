package taskcmd

import (
	"fmt"
	"os"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
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
		task, err := repo.GetTaskByID(id)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), board.FormatTaskDetail(task))
		return err
	},
}

func init() {
	taskCmd.AddCommand(showCmd)
}
