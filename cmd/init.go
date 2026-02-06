package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/storage"
	"mochi-sticky/internal/templates"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold the storage directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := resolveStorageRoot(workingDir, true)
		if err != nil {
			return err
		}
		repo, err := board.NewRepositoryWithStorage(workingDir, storageRoot)
		if err != nil {
			return err
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := repo.InitStoreContext(ctx); err != nil {
			return err
		}
		cfg, err := storage.LoadConfigFromRoot(storageRoot)
		if err != nil {
			return err
		}
		templatePaths, err := storage.ResolveTemplates(workingDir, storageRoot, cfg)
		if err != nil {
			return err
		}
		if err := templates.SeedDefaultsContext(ctx, storageRoot, templatePaths); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Initialized storage at %s\n", storageRoot)
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
