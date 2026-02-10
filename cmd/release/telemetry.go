package release

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"mochi-sticky/internal/cli"
	releasemetrics "mochi-sticky/internal/release"

	"github.com/spf13/cobra"
)

type telemetryReport struct {
	StorageRoot string                 `json:"storage_root"`
	Summary     releasemetrics.Summary `json:"summary"`
	Runs        []releasemetrics.Run   `json:"runs"`
}

var (
	telemetryLimitFlag  int
	telemetryJSONFlag   bool
	telemetryPrettyFlag bool
)

var releaseTelemetryCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "Summarize release telemetry",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		runs, err := releasemetrics.LoadRuns(storageRoot, telemetryLimitFlag)
		if err != nil {
			return err
		}
		summary := releasemetrics.SummarizeRuns(runs)

		if telemetryJSONFlag {
			report := telemetryReport{
				StorageRoot: storageRoot,
				Summary:     summary,
				Runs:        runs,
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetEscapeHTML(false)
			if telemetryPrettyFlag {
				encoder.SetIndent("", "  ")
			}
			return encoder.Encode(report)
		}

		if len(runs) == 0 {
			_, writeErr := fmt.Fprintln(cmd.OutOrStdout(), "No release telemetry found.")
			return writeErr
		}

		if _, err := fmt.Fprintf(
			cmd.OutOrStdout(),
			"Runs: %d (success: %d, failed: %d, success rate: %.1f%%)\n",
			summary.TotalRuns,
			summary.SuccessRuns,
			summary.FailedRuns,
			summary.SuccessRate,
		); err != nil {
			return err
		}

		if summary.SlowestStage != nil {
			if _, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"Slowest stage: %s (%s) on %s in run %s\n",
				summary.SlowestStage.Name,
				formatMilliseconds(summary.SlowestStage.DurationMS),
				summary.SlowestStage.Platform,
				summary.SlowestStage.RunID,
			); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Recent runs:"); err != nil {
			return err
		}
		for _, run := range runs {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", formatRun(run)); err != nil {
				return err
			}
		}
		return nil
	},
}

var releaseTelemetryImportCmd = &cobra.Command{
	Use:   "import <file...>",
	Short: "Import telemetry JSON files into local storage",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, true)
		if err != nil {
			return err
		}
		for _, filePath := range args {
			data, readErr := os.ReadFile(filePath)
			if readErr != nil {
				return fmt.Errorf("read %s: %w", filePath, readErr)
			}
			run, decodeErr := releasemetrics.DecodeRun(data)
			if decodeErr != nil {
				return fmt.Errorf("decode %s: %w", filePath, decodeErr)
			}
			savedPath, saveErr := releasemetrics.SaveRun(storageRoot, run)
			if saveErr != nil {
				return saveErr
			}
			if _, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"Imported telemetry run %s -> %s\n",
				run.RunID,
				savedPath,
			); err != nil {
				return err
			}
		}
		return nil
	},
}

func formatRun(run releasemetrics.Run) string {
	timestamp := run.PrimaryTimestamp().UTC().Format(time.RFC3339)
	status := run.NormalizedStatus()
	if status == "" {
		status = "unknown"
	}
	jobLabel := strings.TrimSpace(run.Job)
	if jobLabel == "" {
		jobLabel = "job"
	}
	platforms := make([]string, 0, len(run.Platforms))
	for _, platform := range run.Platforms {
		if strings.TrimSpace(platform.Name) == "" {
			continue
		}
		platformStatus := strings.TrimSpace(platform.Status)
		if platformStatus == "" {
			platformStatus = "unknown"
		}
		platformDuration := int64(0)
		for _, stage := range platform.Stages {
			platformDuration += stage.DurationMS
		}
		label := fmt.Sprintf("%s:%s", platform.Name, platformStatus)
		if platformDuration > 0 {
			label += fmt.Sprintf(" (%s)", formatMilliseconds(platformDuration))
		}
		platforms = append(platforms, label)
	}
	platformSummary := "no-platform-data"
	if len(platforms) > 0 {
		platformSummary = strings.Join(platforms, ", ")
	}
	return fmt.Sprintf("%s | %s | %s | %s", timestamp, status, jobLabel, platformSummary)
}

func formatMilliseconds(value int64) string {
	if value <= 0 {
		return "0s"
	}
	duration := time.Duration(value) * time.Millisecond
	if duration < time.Second {
		return fmt.Sprintf("%dms", value)
	}
	return duration.Round(time.Millisecond).String()
}

func init() {
	releaseCmd.AddCommand(releaseTelemetryCmd)
	releaseTelemetryCmd.AddCommand(releaseTelemetryImportCmd)
	releaseTelemetryCmd.Flags().IntVar(&telemetryLimitFlag, "limit", 10, "Number of most recent runs to show (0 = all)")
	releaseTelemetryCmd.Flags().BoolVar(&telemetryJSONFlag, "json", false, "Output telemetry report as JSON")
	releaseTelemetryCmd.Flags().BoolVar(&telemetryPrettyFlag, "pretty", false, "Pretty-print JSON output (requires --json)")
}
