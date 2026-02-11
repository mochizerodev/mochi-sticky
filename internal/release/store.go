package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DecodeRun decodes telemetry JSON into a normalized Run.
func DecodeRun(data []byte) (Run, error) {
	var run Run
	if err := json.Unmarshal(data, &run); err != nil {
		return Run{}, fmt.Errorf("release telemetry: failed to decode run: %w", err)
	}
	run = normalizeRun(run)
	if err := validateRun(run); err != nil {
		return Run{}, err
	}
	return run, nil
}

// SaveRun stores a telemetry run under the storage root.
func SaveRun(storageRoot string, run Run) (string, error) {
	if strings.TrimSpace(storageRoot) == "" {
		return "", fmt.Errorf("release telemetry: storage root is required")
	}
	normalized := normalizeRun(run)
	if strings.TrimSpace(normalized.RunID) == "" {
		normalized.RunID = fmt.Sprintf("run-%d", time.Now().UTC().UnixNano())
	}
	if normalized.StartedAt.IsZero() {
		normalized.StartedAt = time.Now().UTC()
	}
	if normalized.FinishedAt.IsZero() {
		normalized.FinishedAt = normalized.StartedAt
	}
	if err := validateRun(normalized); err != nil {
		return "", err
	}

	runsDir := RunsDir(storageRoot)
	if err := os.MkdirAll(runsDir, 0o755); err != nil {
		return "", fmt.Errorf("release telemetry: failed to create runs directory %s: %w", runsDir, err)
	}

	baseName := fileNameForRun(normalized)
	path := filepath.Join(runsDir, baseName)
	for index := 1; ; index++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		} else if err != nil {
			return "", fmt.Errorf("release telemetry: failed to stat %s: %w", path, err)
		}
		path = filepath.Join(runsDir, strings.TrimSuffix(baseName, ".json")+fmt.Sprintf("-%d.json", index))
	}

	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return "", fmt.Errorf("release telemetry: failed to encode run: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("release telemetry: failed to write run %s: %w", path, err)
	}
	return path, nil
}

// LoadRuns loads telemetry runs from storage root, sorted by latest first.
func LoadRuns(storageRoot string, limit int) ([]Run, error) {
	if strings.TrimSpace(storageRoot) == "" {
		return nil, fmt.Errorf("release telemetry: storage root is required")
	}
	runsDir := RunsDir(storageRoot)
	return LoadRunsFromDir(runsDir, limit)
}

// LoadRunsFromDir loads telemetry runs from a specific directory.
func LoadRunsFromDir(runsDir string, limit int) ([]Run, error) {
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("release telemetry: failed to read runs directory %s: %w", runsDir, err)
	}

	runs := make([]Run, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(runsDir, entry.Name())
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("release telemetry: failed to read run %s: %w", path, readErr)
		}
		run, decodeErr := DecodeRun(data)
		if decodeErr != nil {
			return nil, fmt.Errorf("release telemetry: %s: %w", path, decodeErr)
		}
		runs = append(runs, run)
	}

	sort.SliceStable(runs, func(left, right int) bool {
		leftTimestamp := runs[left].PrimaryTimestamp()
		rightTimestamp := runs[right].PrimaryTimestamp()
		if leftTimestamp.Equal(rightTimestamp) {
			return runs[left].RunID > runs[right].RunID
		}
		return leftTimestamp.After(rightTimestamp)
	})

	if limit > 0 && len(runs) > limit {
		return runs[:limit], nil
	}
	return runs, nil
}

func fileNameForRun(run Run) string {
	timestamp := run.PrimaryTimestamp().UTC().Format("20060102T150405Z")
	runID := sanitizeRunID(run.RunID)
	if runID == "" {
		runID = "run"
	}
	return fmt.Sprintf("%s-%s.json", timestamp, runID)
}

func sanitizeRunID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	var builder strings.Builder
	for _, character := range trimmed {
		switch {
		case character >= 'a' && character <= 'z':
			builder.WriteRune(character)
		case character >= 'A' && character <= 'Z':
			builder.WriteRune(character)
		case character >= '0' && character <= '9':
			builder.WriteRune(character)
		case character == '-' || character == '_' || character == '.':
			builder.WriteRune(character)
		default:
			builder.WriteRune('-')
		}
	}
	return builder.String()
}
