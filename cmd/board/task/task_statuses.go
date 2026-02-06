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

var statusesCmd = &cobra.Command{
	Use:   "statuses",
	Short: "List available status keys",
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
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		config, err := repo.LoadConfigContext(ctx)
		if err != nil {
			return err
		}
		if len(config.Columns) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No statuses found.")
			return err
		}

		var b strings.Builder
		for _, column := range config.Columns {
			if strings.TrimSpace(column.Key) == "" {
				continue
			}
			if column.Title != "" {
				fmt.Fprintf(&b, "%s (%s)\n", column.Key, column.Title)
				continue
			}
			fmt.Fprintf(&b, "%s\n", column.Key)
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), b.String())
		return err
	},
}

func init() {
	taskCmd.AddCommand(statusesCmd)
}
