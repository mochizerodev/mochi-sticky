package adr

import "github.com/spf13/cobra"

var adrCmd = &cobra.Command{
	Use:   "adr",
	Short: "Manage architecture decision records",
}

// Register attaches ADR commands to the root command.
func Register(root *cobra.Command) {
	root.AddCommand(adrCmd)
}
