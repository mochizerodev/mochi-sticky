package cmd

import (
	"fmt"

	"mochi-sticky/internal/version"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Version = version.String()
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), version.String())
		return err
	},
}
