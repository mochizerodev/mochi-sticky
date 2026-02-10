package release

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	SchemaVersion = 1
)

// Run captures telemetry for a single workflow/job execution.
type Run struct {
	SchemaVersion int        `json:"schema_version"`
	RunID         string     `json:"run_id"`
	Source        string     `json:"source"`
	Workflow      string     `json:"workflow,omitempty"`
	Job           string     `json:"job,omitempty"`
	Status        string     `json:"status,omitempty"`
	Repository    string     `json:"repository,omitempty"`
	RunURL        string     `json:"run_url,omitempty"`
	RefName       string     `json:"ref_name,omitempty"`
	RefType       string     `json:"ref_type,omitempty"`
	Branch        string     `json:"branch,omitempty"`
	Tag           string     `json:"tag,omitempty"`
	Commit        string     `json:"commit,omitempty"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    time.Time  `json:"finished_at"`
	Config        RunConfig  `json:"config,omitempty"`
	Platforms     []Platform `json:"platforms,omitempty"`
}

// RunConfig captures execution configuration used by a release run.
type RunConfig struct {
	GoVersion     string `json:"go_version,omitempty"`
	GoVersionFile string `json:"go_version_file,omitempty"`
}

// Platform captures telemetry for one platform target within a run.
type Platform struct {
	Name          string  `json:"name"`
	RunnerOS      string  `json:"runner_os,omitempty"`
	RunnerArch    string  `json:"runner_arch,omitempty"`
	Status        string  `json:"status,omitempty"`
	ArtifactBytes int64   `json:"artifact_bytes,omitempty"`
	Stages        []Stage `json:"stages,omitempty"`
}

// Stage captures telemetry for a pipeline stage on a platform.
type Stage struct {
	Name       string `json:"name"`
	Status     string `json:"status,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty"`
}

// RunsDir returns the release telemetry storage directory under the storage root.
func RunsDir(storageRoot string) string {
	return filepath.Join(storageRoot, "release", "telemetry", "runs")
}

// PrimaryTimestamp returns FinishedAt when present, otherwise StartedAt.
func (run Run) PrimaryTimestamp() time.Time {
	if !run.FinishedAt.IsZero() {
		return run.FinishedAt
	}
	return run.StartedAt
}

// NormalizedStatus returns a normalized status value.
func (run Run) NormalizedStatus() string {
	return normalizeStatus(run.Status)
}

func normalizeStatus(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "ok", "pass", "passed", "success", "succeeded":
		return "success"
	case "error", "fail", "failed", "failure":
		return "failure"
	case "cancel", "cancelled", "canceled":
		return "canceled"
	case "skip", "skipped":
		return "skipped"
	case "":
		return ""
	default:
		return normalized
	}
}

func isSuccessStatus(value string) bool {
	return normalizeStatus(value) == "success"
}

func derivePlatformStatus(platform Platform) string {
	if len(platform.Stages) == 0 {
		return normalizeStatus(platform.Status)
	}
	stageStatus := "success"
	for _, stage := range platform.Stages {
		normalized := normalizeStatus(stage.Status)
		if normalized == "" {
			continue
		}
		if normalized == "failure" {
			return "failure"
		}
		if normalized != "success" && stageStatus == "success" {
			stageStatus = normalized
		}
	}
	return stageStatus
}

func deriveRunStatus(run Run) string {
	if len(run.Platforms) == 0 {
		return normalizeStatus(run.Status)
	}
	runStatus := "success"
	for _, platform := range run.Platforms {
		normalized := normalizeStatus(platform.Status)
		if normalized == "" {
			normalized = derivePlatformStatus(platform)
		}
		if normalized == "" {
			continue
		}
		if normalized == "failure" {
			return "failure"
		}
		if normalized != "success" && runStatus == "success" {
			runStatus = normalized
		}
	}
	return runStatus
}

func normalizeRun(run Run) Run {
	normalized := run
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = SchemaVersion
	}
	normalized.RunID = strings.TrimSpace(normalized.RunID)
	normalized.Source = strings.TrimSpace(normalized.Source)
	normalized.Workflow = strings.TrimSpace(normalized.Workflow)
	normalized.Job = strings.TrimSpace(normalized.Job)
	normalized.Repository = strings.TrimSpace(normalized.Repository)
	normalized.RunURL = strings.TrimSpace(normalized.RunURL)
	normalized.RefName = strings.TrimSpace(normalized.RefName)
	normalized.RefType = strings.TrimSpace(normalized.RefType)
	normalized.Branch = strings.TrimSpace(normalized.Branch)
	normalized.Tag = strings.TrimSpace(normalized.Tag)
	normalized.Commit = strings.TrimSpace(normalized.Commit)
	normalized.Config.GoVersion = strings.TrimSpace(normalized.Config.GoVersion)
	normalized.Config.GoVersionFile = strings.TrimSpace(normalized.Config.GoVersionFile)

	normalized.Status = normalizeStatus(normalized.Status)
	if len(normalized.Platforms) > 0 {
		platforms := make([]Platform, 0, len(normalized.Platforms))
		for _, platform := range normalized.Platforms {
			normalizedPlatform := platform
			normalizedPlatform.Name = strings.TrimSpace(normalizedPlatform.Name)
			normalizedPlatform.RunnerOS = strings.TrimSpace(normalizedPlatform.RunnerOS)
			normalizedPlatform.RunnerArch = strings.TrimSpace(normalizedPlatform.RunnerArch)
			normalizedPlatform.Status = normalizeStatus(normalizedPlatform.Status)
			if len(normalizedPlatform.Stages) > 0 {
				stages := make([]Stage, 0, len(normalizedPlatform.Stages))
				for _, stage := range normalizedPlatform.Stages {
					normalizedStage := stage
					normalizedStage.Name = strings.TrimSpace(normalizedStage.Name)
					normalizedStage.Status = normalizeStatus(normalizedStage.Status)
					if normalizedStage.DurationMS < 0 {
						normalizedStage.DurationMS = 0
					}
					stages = append(stages, normalizedStage)
				}
				normalizedPlatform.Stages = stages
			}
			if normalizedPlatform.Status == "" {
				normalizedPlatform.Status = derivePlatformStatus(normalizedPlatform)
			}
			platforms = append(platforms, normalizedPlatform)
		}
		normalized.Platforms = platforms
	}
	if normalized.Status == "" {
		normalized.Status = deriveRunStatus(normalized)
	}
	return normalized
}

func validateRun(run Run) error {
	if strings.TrimSpace(run.RunID) == "" {
		return fmt.Errorf("release telemetry: run_id is required")
	}
	if run.StartedAt.IsZero() {
		return fmt.Errorf("release telemetry: started_at is required")
	}
	if run.FinishedAt.IsZero() {
		return fmt.Errorf("release telemetry: finished_at is required")
	}
	if run.FinishedAt.Before(run.StartedAt) {
		return fmt.Errorf("release telemetry: finished_at must be >= started_at")
	}
	for index, platform := range run.Platforms {
		if strings.TrimSpace(platform.Name) == "" {
			return fmt.Errorf("release telemetry: platforms[%d].name is required", index)
		}
		for stageIndex, stage := range platform.Stages {
			if strings.TrimSpace(stage.Name) == "" {
				return fmt.Errorf("release telemetry: platforms[%d].stages[%d].name is required", index, stageIndex)
			}
		}
	}
	return nil
}
