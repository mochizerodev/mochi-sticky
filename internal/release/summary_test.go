package release

import (
	"testing"
	"time"
)

func TestSummarizeRuns(t *testing.T) {
	runs := []Run{
		{
			RunID:      "run-1",
			Status:     "success",
			StartedAt:  time.Date(2026, 2, 10, 11, 0, 0, 0, time.UTC),
			FinishedAt: time.Date(2026, 2, 10, 11, 5, 0, 0, time.UTC),
			Platforms: []Platform{
				{
					Name:   "linux-amd64",
					Status: "success",
					Stages: []Stage{
						{Name: "test", Status: "success", DurationMS: 15000},
						{Name: "build", Status: "success", DurationMS: 25000},
					},
				},
			},
		},
		{
			RunID:      "run-2",
			Status:     "failure",
			StartedAt:  time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
			FinishedAt: time.Date(2026, 2, 10, 12, 7, 0, 0, time.UTC),
			Platforms: []Platform{
				{
					Name:   "linux-amd64",
					Status: "failure",
					Stages: []Stage{
						{Name: "test", Status: "success", DurationMS: 12000},
						{Name: "build", Status: "failure", DurationMS: 30000},
					},
				},
				{
					Name:   "windows-amd64",
					Status: "success",
					Stages: []Stage{
						{Name: "smoke", Status: "success", DurationMS: 18000},
					},
				},
			},
		},
	}

	summary := SummarizeRuns(runs)
	if summary.TotalRuns != 2 {
		t.Fatalf("expected total runs 2, got %d", summary.TotalRuns)
	}
	if summary.SuccessRuns != 1 {
		t.Fatalf("expected success runs 1, got %d", summary.SuccessRuns)
	}
	if summary.FailedRuns != 1 {
		t.Fatalf("expected failed runs 1, got %d", summary.FailedRuns)
	}
	if summary.SuccessRate != 50.0 {
		t.Fatalf("expected success rate 50.0, got %.1f", summary.SuccessRate)
	}
	if summary.SlowestStage == nil {
		t.Fatalf("expected slowest stage")
	}
	if summary.SlowestStage.Name != "build" || summary.SlowestStage.DurationMS != 30000 {
		t.Fatalf("unexpected slowest stage: %+v", summary.SlowestStage)
	}
	if len(summary.Platforms) != 2 {
		t.Fatalf("expected 2 platform summaries, got %d", len(summary.Platforms))
	}
}
