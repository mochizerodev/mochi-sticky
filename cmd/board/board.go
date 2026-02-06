package board

import (
	"github.com/spf13/cobra"
)

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Manage boards",
}

func Register(root *cobra.Command) {
	root.AddCommand(boardCmd)
}
