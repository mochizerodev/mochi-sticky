package wiki

import "github.com/spf13/cobra"

var wikiCmd = &cobra.Command{
	Use:   "wiki",
	Short: "Manage wiki pages",
}

func Register(root *cobra.Command) {
	root.AddCommand(wikiCmd)
}
