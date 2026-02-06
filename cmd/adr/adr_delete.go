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

var adrDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an ADR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := adr.ParseID(args[0])
		if err != nil {
			return err
		}
		if err := cli.RequireConfirm(cmd, fmt.Sprintf("Delete ADR %s? This cannot be undone.", adr.FormatID(id))); err != nil {
			return err
		}
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
		if err := repo.DeleteADRContext(ctx, id); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted ADR %s\n", adr.FormatID(id))
		return err
	},
}

func init() {
	adrDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	adrCmd.AddCommand(adrDeleteCmd)
}
