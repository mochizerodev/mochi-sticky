package release

import (
	"sort"
)

// SlowestStage captures the slowest measured stage.
type SlowestStage struct {
	RunID      string `json:"run_id"`
	Platform   string `json:"platform"`
	Name       string `json:"name"`
	DurationMS int64  `json:"duration_ms"`
}

// PlatformSummary captures aggregate run outcomes for one platform.
type PlatformSummary struct {
	Platform    string  `json:"platform"`
	TotalRuns   int     `json:"total_runs"`
	SuccessRuns int     `json:"success_runs"`
	FailedRuns  int     `json:"failed_runs"`
	SuccessRate float64 `json:"success_rate"`
}

// Summary captures aggregate telemetry metrics across runs.
type Summary struct {
	TotalRuns    int               `json:"total_runs"`
	SuccessRuns  int               `json:"success_runs"`
	FailedRuns   int               `json:"failed_runs"`
	SuccessRate  float64           `json:"success_rate"`
	SlowestStage *SlowestStage     `json:"slowest_stage,omitempty"`
	Platforms    []PlatformSummary `json:"platforms,omitempty"`
}

// SummarizeRuns computes aggregate metrics used by CLI/TUI dashboards.
func SummarizeRuns(runs []Run) Summary {
	summary := Summary{
		TotalRuns: len(runs),
	}
	if len(runs) == 0 {
		return summary
	}

	platforms := map[string]*PlatformSummary{}
	for _, run := range runs {
		if isSuccessStatus(run.Status) {
			summary.SuccessRuns++
		} else {
			summary.FailedRuns++
		}

		for _, platform := range run.Platforms {
			entry, ok := platforms[platform.Name]
			if !ok {
				entry = &PlatformSummary{Platform: platform.Name}
				platforms[platform.Name] = entry
			}
			entry.TotalRuns++
			if isSuccessStatus(platform.Status) {
				entry.SuccessRuns++
			} else {
				entry.FailedRuns++
			}
			for _, stage := range platform.Stages {
				if stage.DurationMS <= 0 {
					continue
				}
				if summary.SlowestStage == nil || stage.DurationMS > summary.SlowestStage.DurationMS {
					summary.SlowestStage = &SlowestStage{
						RunID:      run.RunID,
						Platform:   platform.Name,
						Name:       stage.Name,
						DurationMS: stage.DurationMS,
					}
				}
			}
		}
	}

	summary.SuccessRate = roundSuccessRate(summary.SuccessRuns, summary.TotalRuns)

	if len(platforms) > 0 {
		list := make([]PlatformSummary, 0, len(platforms))
		for _, entry := range platforms {
			entry.SuccessRate = roundSuccessRate(entry.SuccessRuns, entry.TotalRuns)
			list = append(list, *entry)
		}
		sort.SliceStable(list, func(left, right int) bool {
			return list[left].Platform < list[right].Platform
		})
		summary.Platforms = list
	}

	return summary
}

func roundSuccessRate(success, total int) float64 {
	if total <= 0 {
		return 0
	}
	rate := (float64(success) / float64(total)) * 100
	return float64(int(rate*10+0.5)) / 10
}
