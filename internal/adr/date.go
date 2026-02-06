package adr

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Date wraps time.Time so ADR YAML frontmatter round-trips using YYYY-MM-DD.
type Date struct {
	time.Time
}

// UnmarshalYAML parses a YYYY-MM-DD date string.
func (d *Date) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value.Value)
	if err != nil {
		return fmt.Errorf("adr: failed to parse date %q: %w", value.Value, err)
	}
	d.Time = parsed
	return nil
}

// MarshalYAML formats the date as YYYY-MM-DD for the YAML frontmatter.
func (d Date) MarshalYAML() (any, error) {
	if d.Time.IsZero() {
		return "", nil
	}
	return d.Format("2006-01-02"), nil
}

// IsZero reports whether the date is unset.
func (d Date) IsZero() bool {
	return d.Time.IsZero()
}
