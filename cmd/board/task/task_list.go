package taskcmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
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
		tasks, err := repo.GetAllTasks()
		if err != nil {
			return err
		}
		statusFilter, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}
		titleFilter, err := cmd.Flags().GetString("title")
		if err != nil {
			return err
		}
		tagFilters, err := cmd.Flags().GetStringSlice("tag")
		if err != nil {
			return err
		}
		tagMode, err := cmd.Flags().GetString("tag-mode")
		if err != nil {
			return err
		}
		fromStr, err := cmd.Flags().GetString("from")
		if err != nil {
			return err
		}
		toStr, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}
		var fromDate time.Time
		if strings.TrimSpace(fromStr) != "" {
			parsed, err := time.Parse("2006-01-02", fromStr)
			if err != nil {
				return err
			}
			fromDate = parsed
		}
		var toDate time.Time
		if strings.TrimSpace(toStr) != "" {
			parsed, err := time.Parse("2006-01-02", toStr)
			if err != nil {
				return err
			}
			toDate = parsed
		}
		sortBy, err := cmd.Flags().GetString("sort")
		if err != nil {
			return err
		}
		desc, err := cmd.Flags().GetBool("desc")
		if err != nil {
			return err
		}

		tasks = board.FilterAndSortTasks(tasks, board.ListOptions{
			Status:  statusFilter,
			Title:   titleFilter,
			Tags:    board.NormalizeTags(tagFilters),
			TagMode: tagMode,
			From:    fromDate,
			To:      toDate,
			SortBy:  sortBy,
			Desc:    desc,
		})

		if len(tasks) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No tasks found.")
			return err
		}
		table := board.FormatTasksTable(tasks)
		_, err = fmt.Fprintln(cmd.OutOrStdout(), table)
		return err
	},
}

func init() {
	taskCmd.AddCommand(listCmd)
	listCmd.Flags().String("status", "", "Filter tasks by status key")
	listCmd.Flags().String("title", "", "Filter tasks by title (substring match)")
	listCmd.Flags().StringSlice("tag", nil, "Filter tasks by tag (repeatable)")
	listCmd.Flags().String("tag-mode", "any", "Tag match mode: any|all")
	listCmd.Flags().String("from", "", "Filter tasks created on/after YYYY-MM-DD")
	listCmd.Flags().String("to", "", "Filter tasks created on/before YYYY-MM-DD")
	listCmd.Flags().String("sort", "", "Sort by: status, created, title, priority")
	listCmd.Flags().Bool("desc", false, "Sort in descending order")
}
