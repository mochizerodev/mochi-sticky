package adr

import (
	"fmt"
	"strings"
)

var requiredHeadings = []string{
	"## Context",
	"## Decision",
	"## Consequences",
}

// ValidateADR performs lightweight validation of frontmatter and required headings.
func ValidateADR(adr ADR) error {
	if strings.TrimSpace(adr.Title) == "" {
		return fmt.Errorf("adr: %w", ErrInvalidTitle)
	}
	if adr.ID <= 0 {
		return fmt.Errorf("adr: %w", ErrInvalidID)
	}
	if err := ValidateRequiredHeadings(adr.Content); err != nil {
		return err
	}
	return nil
}

// ValidateRequiredHeadings ensures the ADR body includes required top-level headings.
func ValidateRequiredHeadings(content string) error {
	found := make(map[string]bool, len(requiredHeadings))
	for _, heading := range requiredHeadings {
		found[heading] = false
	}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, heading := range requiredHeadings {
			if strings.EqualFold(trimmed, heading) {
				found[heading] = true
			}
		}
	}
	var missing []string
	for _, heading := range requiredHeadings {
		if !found[heading] {
			missing = append(missing, heading)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("adr: missing required section(s): %s", strings.Join(missing, ", "))
	}
	return nil
}
