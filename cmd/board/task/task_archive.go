package taskcmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Manage archived tasks",
}

var archiveTaskCmd = &cobra.Command{
	Use:   "task <id>",
	Short: "Archive a task by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cli.RequireConfirm(cmd, fmt.Sprintf("Archive task %q?", args[0])); err != nil {
			return err
		}
		repo, err := cli.RepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		task, err := repo.ArchiveTaskContext(ctx, args[0])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Archived task %s\n", task.ID)
		return err
	},
}

var archiveBeforeCmd = &cobra.Command{
	Use:   "before <YYYY-MM-DD>",
	Short: "Archive tasks created before a date",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cli.RequireConfirm(cmd, fmt.Sprintf("Archive tasks before %s?", args[0])); err != nil {
			return err
		}
		cutoff, err := time.Parse("2006-01-02", args[0])
		if err != nil {
			return err
		}
		repo, err := cli.RepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		moved, err := repo.ArchiveBeforeContext(ctx, cutoff)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Archived %d tasks\n", len(moved))
		return err
	},
}

var archiveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List archived tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := cli.RepoFromCwd()
		if err != nil {
			return err
		}
		tasks, err := repo.ListArchivedTasks()
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No archived tasks found.")
			return err
		}
		table := board.FormatTasksTable(tasks)
		_, err = fmt.Fprintln(cmd.OutOrStdout(), table)
		return err
	},
}

var archiveRestoreCmd = &cobra.Command{
	Use:   "restore <id>",
	Short: "Restore an archived task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cli.RequireConfirm(cmd, fmt.Sprintf("Restore task %q?", args[0])); err != nil {
			return err
		}
		repo, err := cli.RepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		task, err := repo.RestoreTaskContext(ctx, args[0])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Restored task %s\n", task.ID)
		return err
	},
}

var archiveDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an archived task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		message := fmt.Sprintf("Delete archived task %q? This cannot be undone.", args[0])
		if err := cli.RequireConfirm(cmd, message); err != nil {
			return err
		}
		repo, err := cli.RepoFromCwd()
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.DeleteArchivedTaskContext(ctx, args[0]); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted archived task %s\n", args[0])
		return err
	},
}

func init() {
	archiveTaskCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	archiveBeforeCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	archiveRestoreCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	archiveDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	archiveCmd.AddCommand(archiveTaskCmd)
	archiveCmd.AddCommand(archiveBeforeCmd)
	archiveCmd.AddCommand(archiveListCmd)
	archiveCmd.AddCommand(archiveRestoreCmd)
	archiveCmd.AddCommand(archiveDeleteCmd)
	taskCmd.AddCommand(archiveCmd)
}
