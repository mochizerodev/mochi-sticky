package adr

import (
	"fmt"
	"os"
	"strings"

	"mochi-sticky/internal/adr"
	"mochi-sticky/internal/board"
	"mochi-sticky/internal/cli"

	"github.com/spf13/cobra"
)

var adrListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ADRs",
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
		adrs, err := repo.ListADRs()
		if err != nil {
			return err
		}

		status, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}
		tagsRaw, err := cmd.Flags().GetString("tags")
		if err != nil {
			return err
		}
		query, err := cmd.Flags().GetString("query")
		if err != nil {
			return err
		}
		sinceRaw, err := cmd.Flags().GetString("since")
		if err != nil {
			return err
		}
		untilRaw, err := cmd.Flags().GetString("until")
		if err != nil {
			return err
		}
		since, err := parseDate(sinceRaw)
		if err != nil {
			return err
		}
		until, err := parseDate(untilRaw)
		if err != nil {
			return err
		}

		adrs = adr.FilterADRs(adrs, adr.FilterOptions{
			Status:          status,
			Tags:            board.ParseTags(tagsRaw),
			Query:           query,
			Since:           since,
			Until:           until,
			CaseInsensitive: true,
		})
		adr.SortADRs(adrs)
		if len(adrs) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No ADRs found.")
			return err
		}

		for _, record := range adrs {
			statusLabel := strings.TrimSpace(record.Status)
			if statusLabel == "" {
				statusLabel = "unknown"
			}
			dateLabel := "-"
			if !record.Date.IsZero() {
				dateLabel = record.Date.Format("2006-01-02")
			}
			if _, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"%s\t%s\t%s\t%s\n",
				adr.FormatID(record.ID),
				statusLabel,
				dateLabel,
				record.Title,
			); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	adrCmd.AddCommand(adrListCmd)
	adrListCmd.Flags().String("status", "", "Filter by status key")
	adrListCmd.Flags().String("tags", "", "Filter by tags (comma-separated)")
	adrListCmd.Flags().String("query", "", "Filter by keyword query (title/body)")
	adrListCmd.Flags().String("since", "", "Filter by date >= YYYY-MM-DD")
	adrListCmd.Flags().String("until", "", "Filter by date <= YYYY-MM-DD")
}
