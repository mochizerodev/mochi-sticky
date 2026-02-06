package adr

import (
	"strings"
	"testing"
	"time"
)

func TestRenderParseADR_RoundTrip(t *testing.T) {
	original := ADR{
		ID:      1,
		UID:     "00000000-0000-0000-0000-000000000000",
		Title:   "Use YAML Frontmatter",
		Status:  "proposed",
		Date:    Date{Time: time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC)},
		Tags:    []string{"architecture", "format"},
		Links:   []string{"T-000007"},
		Content: DefaultContent(),
	}

	data, err := RenderADR(original)
	if err != nil {
		t.Fatalf("RenderADR: %v", err)
	}
	parsed, err := ParseADR(data)
	if err != nil {
		t.Fatalf("ParseADR: %v", err)
	}
	if parsed.ID != original.ID {
		t.Fatalf("expected id %d, got %d", original.ID, parsed.ID)
	}
	if parsed.Title != original.Title {
		t.Fatalf("expected title %q, got %q", original.Title, parsed.Title)
	}
	if parsed.Status != original.Status {
		t.Fatalf("expected status %q, got %q", original.Status, parsed.Status)
	}
	if parsed.Date.Format("2006-01-02") != original.Date.Format("2006-01-02") {
		t.Fatalf("expected date %s, got %s", original.Date.Format("2006-01-02"), parsed.Date.Format("2006-01-02"))
	}
	if strings.TrimSpace(parsed.Content) != strings.TrimSpace(original.Content) {
		t.Fatalf("expected content to round trip")
	}
}

func TestValidateRequiredHeadings(t *testing.T) {
	if err := ValidateRequiredHeadings(DefaultContent()); err != nil {
		t.Fatalf("expected default content valid, got %v", err)
	}
	if err := ValidateRequiredHeadings("## Context\n\n"); err == nil {
		t.Fatalf("expected missing headings error")
	}
}
