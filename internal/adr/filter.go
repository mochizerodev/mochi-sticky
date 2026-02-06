package adr

import (
	"sort"
	"strings"
	"time"
)

// FilterOptions defines simple list filtering for ADRs.
type FilterOptions struct {
	Status          string
	Tags            []string
	Query           string
	Since           time.Time
	Until           time.Time
	CaseInsensitive bool
}

// FilterADRs returns ADRs matching the provided filter options.
func FilterADRs(adrs []ADR, opts FilterOptions) []ADR {
	status := strings.TrimSpace(opts.Status)
	query := strings.TrimSpace(opts.Query)
	if opts.CaseInsensitive {
		status = strings.ToLower(status)
		query = strings.ToLower(query)
	}

	tagSet := make(map[string]struct{}, len(opts.Tags))
	for _, tag := range opts.Tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if opts.CaseInsensitive {
			trimmed = strings.ToLower(trimmed)
		}
		tagSet[trimmed] = struct{}{}
	}

	out := make([]ADR, 0, len(adrs))
	for _, adr := range adrs {
		if status != "" {
			value := strings.TrimSpace(adr.Status)
			if opts.CaseInsensitive {
				value = strings.ToLower(value)
			}
			if value != status {
				continue
			}
		}
		if len(tagSet) > 0 {
			matched := false
			for _, tag := range adr.Tags {
				t := strings.TrimSpace(tag)
				if opts.CaseInsensitive {
					t = strings.ToLower(t)
				}
				if _, ok := tagSet[t]; ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if !opts.Since.IsZero() || !opts.Until.IsZero() {
			if adr.Date.IsZero() {
				continue
			}
			when := adr.Date.Time
			if !opts.Since.IsZero() && when.Before(opts.Since) {
				continue
			}
			if !opts.Until.IsZero() && when.After(opts.Until) {
				continue
			}
		}
		if query != "" {
			title := strings.TrimSpace(adr.Title)
			body := strings.TrimSpace(adr.Content)
			if opts.CaseInsensitive {
				title = strings.ToLower(title)
				body = strings.ToLower(body)
			}
			if !strings.Contains(title, query) && !strings.Contains(body, query) {
				continue
			}
		}
		out = append(out, adr)
	}
	return out
}

// SortADRs sorts ADRs newest-first (by Date then ID).
func SortADRs(adrs []ADR) {
	sort.Slice(adrs, func(i, j int) bool {
		left := adrs[i]
		right := adrs[j]
		if !left.Date.IsZero() || !right.Date.IsZero() {
			if left.Date.IsZero() {
				return false
			}
			if right.Date.IsZero() {
				return true
			}
			if !left.Date.Equal(right.Date.Time) {
				return left.Date.After(right.Date.Time)
			}
		}
		if left.ID != right.ID {
			return left.ID > right.ID
		}
		return strings.ToLower(left.Title) < strings.ToLower(right.Title)
	})
}
