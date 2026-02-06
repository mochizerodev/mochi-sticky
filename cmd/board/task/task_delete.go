package taskcmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cli.RequireConfirm(cmd, fmt.Sprintf("Delete task %q? This cannot be undone.", args[0])); err != nil {
			return err
		}
		repo, err := cli.RepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.DeleteTaskContext(ctx, args[0]); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted task %s\n", args[0])
		return err
	},
}

func init() {
	deleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	taskCmd.AddCommand(deleteCmd)
}
