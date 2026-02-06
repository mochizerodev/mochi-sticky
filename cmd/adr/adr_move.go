package adr

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var adrMoveCmd = &cobra.Command{
	Use:   "move <id> <status>",
	Short: "Move an ADR to a new status",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := adr.ParseID(args[0])
		if err != nil {
			return err
		}
		status := args[1]

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
		if err := repo.UpdateADRStatusContext(ctx, id, status); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Moved ADR %s to %s\n", adr.FormatID(id), status)
		return err
	},
}

func init() {
	adrCmd.AddCommand(adrMoveCmd)
}
