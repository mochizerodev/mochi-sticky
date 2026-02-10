package release

import "github.com/spf13/cobra"

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release workflow helpers",
}

func Register(root *cobra.Command) {
	root.AddCommand(releaseCmd)
}
