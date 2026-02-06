package adr

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var adrStatusesCmd = &cobra.Command{
	Use:   "statuses",
	Short: "List available ADR status keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		root := adrRoot(storageRoot)
		repo, err := adr.NewRepository(root)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.InitStoreContext(ctx); err != nil {
			return err
		}
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
			key := strings.TrimSpace(column.Key)
			if key == "" {
				continue
			}
			title := strings.TrimSpace(column.Title)
			if title != "" {
				fmt.Fprintf(&b, "%s (%s)\n", key, title)
				continue
			}
			fmt.Fprintf(&b, "%s\n", key)
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), b.String())
		return err
	},
}

func init() {
	adrCmd.AddCommand(adrStatusesCmd)
}
