package cmd

import (
	"fmt"
	"os"

	"mochi-sticky/cmd/adr"
	"mochi-sticky/cmd/board"
	taskcmd "mochi-sticky/cmd/board/task"
	"mochi-sticky/cmd/tui"
	"mochi-sticky/cmd/wiki"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mochi-sticky",
	Short: "mochi-sticky is a file-based Kanban board and wiki for developers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&storageRootFlag,
		"storage",
		"",
		"Storage root for boards/wiki/adrs (overrides config/env)",
	)
	adr.Register(rootCmd)
	board.Register(rootCmd)
	taskcmd.Register(rootCmd)
	wiki.Register(rootCmd)
	tui.Register(rootCmd)
}
