package release

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadRunsSortedWithLimit(t *testing.T) {
	storageRoot := t.TempDir()

	runOne := Run{
		RunID:      "1001",
		Source:     "github-actions",
		Status:     "success",
		StartedAt:  time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC),
		FinishedAt: time.Date(2026, 2, 10, 10, 10, 0, 0, time.UTC),
		Platforms: []Platform{
			{
				Name:   "linux-amd64",
				Status: "success",
				Stages: []Stage{
					{Name: "test", Status: "success", DurationMS: 12000},
				},
			},
		},
	}
	runTwo := Run{
		RunID:      "1002",
		Source:     "github-actions",
		Status:     "failure",
		StartedAt:  time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
		FinishedAt: time.Date(2026, 2, 10, 12, 20, 0, 0, time.UTC),
		Platforms: []Platform{
			{
				Name:   "windows-amd64",
				Status: "failure",
				Stages: []Stage{
					{Name: "build", Status: "failure", DurationMS: 40000},
				},
			},
		},
	}

	pathOne, err := SaveRun(storageRoot, runOne)
	if err != nil {
		t.Fatalf("save run one: %v", err)
	}
	if _, err := os.Stat(pathOne); err != nil {
		t.Fatalf("stat run one: %v", err)
	}
	pathTwo, err := SaveRun(storageRoot, runTwo)
	if err != nil {
		t.Fatalf("save run two: %v", err)
	}
	if _, err := os.Stat(pathTwo); err != nil {
		t.Fatalf("stat run two: %v", err)
	}

	allRuns, err := LoadRuns(storageRoot, 0)
	if err != nil {
		t.Fatalf("load runs: %v", err)
	}
	if len(allRuns) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(allRuns))
	}
	if allRuns[0].RunID != "1002" {
		t.Fatalf("expected latest run first, got %s", allRuns[0].RunID)
	}

	limitedRuns, err := LoadRuns(storageRoot, 1)
	if err != nil {
		t.Fatalf("load limited runs: %v", err)
	}
	if len(limitedRuns) != 1 {
		t.Fatalf("expected 1 run, got %d", len(limitedRuns))
	}
	if limitedRuns[0].RunID != "1002" {
		t.Fatalf("expected latest limited run to be 1002, got %s", limitedRuns[0].RunID)
	}
}

func TestDecodeRunNormalizesStatuses(t *testing.T) {
	data := []byte(`{
  "schema_version": 1,
  "run_id": "abc-1",
  "source": "github-actions",
  "started_at": "2026-02-10T13:00:00Z",
  "finished_at": "2026-02-10T13:02:00Z",
  "platforms": [
    {
      "name": "linux-amd64",
      "stages": [
        { "name": "test", "status": "passed", "duration_ms": 2100 },
        { "name": "build", "status": "succeeded", "duration_ms": 4200 }
      ]
    }
  ]
}`)

	run, err := DecodeRun(data)
	if err != nil {
		t.Fatalf("decode run: %v", err)
	}
	if run.Status != "success" {
		t.Fatalf("expected normalized run status success, got %s", run.Status)
	}
	if len(run.Platforms) != 1 {
		t.Fatalf("expected one platform, got %d", len(run.Platforms))
	}
	if run.Platforms[0].Status != "success" {
		t.Fatalf("expected normalized platform status success, got %s", run.Platforms[0].Status)
	}
}

func TestLoadRunsFromMissingDirReturnsEmpty(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	runs, err := LoadRunsFromDir(missing, 0)
	if err != nil {
		t.Fatalf("load runs from missing dir: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected no runs, got %d", len(runs))
	}
}
