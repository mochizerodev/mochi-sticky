package wiki

import (
	"fmt"
	"strings"
)

// LintIssue reports a wiki lint finding.
type LintIssue struct {
	Slug     string `json:"slug"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// LintPages checks pages for basic structure issues.
func LintPages(pages []Page) []LintIssue {
	issues := make([]LintIssue, 0)
	for _, page := range pages {
		slug := page.Slug
		if strings.TrimSpace(slug) == "" {
			slug = "(missing slug)"
		}
		if strings.TrimSpace(page.Title) == "" {
			issues = append(issues, LintIssue{
				Slug:     slug,
				Severity: "error",
				Message:  "missing title",
			})
		}
		if strings.TrimSpace(page.Content) == "" {
			issues = append(issues, LintIssue{
				Slug:     slug,
				Severity: "warning",
				Message:  "empty content",
			})
			continue
		}
		lines := strings.Split(page.Content, "\n")
		if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "#") {
			issues = append(issues, LintIssue{
				Slug:     slug,
				Severity: "warning",
				Message:  "first content line is not a heading",
			})
		}
		if err := lintLinks(page.Content); err != nil {
			issues = append(issues, LintIssue{
				Slug:     slug,
				Severity: "warning",
				Message:  err.Error(),
			})
		}
	}
	return issues
}

func lintLinks(content string) error {
	if !strings.Contains(content, "](") {
		return nil
	}
	bad := 0
	parts := strings.Split(content, "](")
	for i := 1; i < len(parts); i++ {
		rest := parts[i]
		end := strings.Index(rest, ")")
		if end == -1 {
			bad++
			continue
		}
		target := strings.TrimSpace(rest[:end])
		if target == "" {
			bad++
		}
	}
	if bad > 0 {
		return fmt.Errorf("invalid link syntax in %d places", bad)
	}
	return nil
}
