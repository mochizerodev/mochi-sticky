package taskcmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		task, err := board.NewTask(title)
		if err != nil {
			return err
		}
		tagsInput, err := cmd.Flags().GetString("tags")
		if err != nil {
			return err
		}
		if tagsInput != "" {
			task.Tags = board.ParseTags(tagsInput)
		}
		priority, err := cmd.Flags().GetInt("priority")
		if err != nil {
			return err
		}
		task.Priority = priority

		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
		if err != nil {
			return err
		}
		repo, err := board.NewRepositoryWithStorage(workingDir, storageRoot)
		if err != nil {
			return err
		}
		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}
		if strings.TrimSpace(templateName) != "" {
			templateDir := templatePaths.Task
			if strings.TrimSpace(templateDir) == "" {
				templateDir = filepath.Join(storageRoot, "templates", "task")
			}
			templatePath := filepath.Join(templateDir, templateName+".md")
			data, err := os.ReadFile(templatePath)
			if err != nil {
				return fmt.Errorf("template not found: %s", templateName)
			}
			task.Content = string(data)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		created, err := repo.CreateTaskContext(ctx, task)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created task %s\n", created.ID)
		return err
	},
}

func init() {
	taskCmd.AddCommand(addCmd)
	addCmd.Flags().String("tags", "", "Comma-separated tags")
	addCmd.Flags().Int("priority", board.DefaultPriority, "Priority (1-3)")
	addCmd.Flags().String("template", "", "Template name (from configured task templates)")
}
