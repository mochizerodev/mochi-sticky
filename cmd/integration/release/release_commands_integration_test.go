package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	releasemetrics "mochi-sticky/internal/release"
)

func TestReleaseTelemetryImportAndSummary(t *testing.T) {
	repoRoot, storageRoot := setupReleaseStorage(t)

	runOne := releasemetrics.Run{
		SchemaVersion: 1,
		RunID:         "run-100",
		Source:        "github-actions",
		Workflow:      "Release",
		Job:           "build",
		Status:        "success",
		Repository:    "mochizero0/mochi-sticky",
		RefName:       "v0.5.0",
		RefType:       "tag",
		Tag:           "v0.5.0",
		Commit:        "abc123",
		StartedAt:     time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC),
		FinishedAt:    time.Date(2026, 2, 10, 10, 5, 0, 0, time.UTC),
		Platforms: []releasemetrics.Platform{
			{
				Name:          "linux-amd64",
				Status:        "success",
				ArtifactBytes: 1024,
				Stages: []releasemetrics.Stage{
					{Name: "test", Status: "success", DurationMS: 30000},
					{Name: "build", Status: "success", DurationMS: 45000},
					{Name: "export", Status: "success", DurationMS: 5000},
				},
			},
		},
	}
	runTwo := releasemetrics.Run{
		SchemaVersion: 1,
		RunID:         "run-101",
		Source:        "github-actions",
		Workflow:      "Release",
		Job:           "smoke",
		Status:        "failure",
		Repository:    "mochizero0/mochi-sticky",
		RefName:       "v0.5.0",
		RefType:       "tag",
		Tag:           "v0.5.0",
		Commit:        "abc123",
		StartedAt:     time.Date(2026, 2, 10, 11, 0, 0, 0, time.UTC),
		FinishedAt:    time.Date(2026, 2, 10, 11, 6, 0, 0, time.UTC),
		Platforms: []releasemetrics.Platform{
			{
				Name:   "windows-amd64",
				Status: "failure",
				Stages: []releasemetrics.Stage{
					{Name: "build", Status: "success", DurationMS: 25000},
					{Name: "smoke", Status: "failure", DurationMS: 65000},
				},
			},
		},
	}

	inputOne := filepath.Join(t.TempDir(), "run-one.json")
	inputTwo := filepath.Join(t.TempDir(), "run-two.json")
	writeRunJSON(t, inputOne, runOne)
	writeRunJSON(t, inputTwo, runTwo)

	importOut, importErr := runMochiSticky(t, repoRoot, storageRoot, "release", "telemetry", "import", inputOne, inputTwo)
	if importErr != nil {
		t.Fatalf("telemetry import: %v", importErr)
	}
	if !strings.Contains(importOut, "Imported telemetry run run-100") {
		t.Fatalf("expected import output for run-100, got:\n%s", importOut)
	}
	if !strings.Contains(importOut, "Imported telemetry run run-101") {
		t.Fatalf("expected import output for run-101, got:\n%s", importOut)
	}

	summaryOut, summaryErr := runMochiSticky(t, repoRoot, storageRoot, "release", "telemetry")
	if summaryErr != nil {
		t.Fatalf("release telemetry summary: %v", summaryErr)
	}
	if !strings.Contains(summaryOut, "Runs: 2 (success: 1, failed: 1, success rate: 50.0%)") {
		t.Fatalf("unexpected summary output:\n%s", summaryOut)
	}
	if !strings.Contains(summaryOut, "Slowest stage: smoke") {
		t.Fatalf("expected slowest stage in output, got:\n%s", summaryOut)
	}

	jsonOut, jsonErr := runMochiSticky(t, repoRoot, storageRoot, "release", "telemetry", "--json")
	if jsonErr != nil {
		t.Fatalf("release telemetry --json: %v", jsonErr)
	}
	var report struct {
		Summary releasemetrics.Summary `json:"summary"`
		Runs    []releasemetrics.Run   `json:"runs"`
	}
	if err := json.Unmarshal([]byte(jsonOut), &report); err != nil {
		t.Fatalf("parse json report: %v", err)
	}
	if report.Summary.TotalRuns != 2 {
		t.Fatalf("expected total_runs 2, got %d", report.Summary.TotalRuns)
	}
	if report.Summary.SuccessRuns != 1 {
		t.Fatalf("expected success_runs 1, got %d", report.Summary.SuccessRuns)
	}
	if report.Summary.SlowestStage == nil || report.Summary.SlowestStage.Name != "smoke" {
		t.Fatalf("expected slowest stage smoke, got %+v", report.Summary.SlowestStage)
	}
	if len(report.Runs) != 2 {
		t.Fatalf("expected 2 runs in JSON report, got %d", len(report.Runs))
	}
}

func writeRunJSON(t *testing.T, path string, run releasemetrics.Run) {
	t.Helper()

	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("marshal run: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write run json: %v", err)
	}
}
