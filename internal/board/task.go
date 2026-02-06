package board

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultStatus is the initial status for new tasks.
	DefaultStatus = "todo"
	// DefaultPriority is the initial priority for new tasks.
	DefaultPriority = 2
)

// Date wraps time.Time so the YAML frontmatter round-trips using YYYY-MM-DD.
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
		return fmt.Errorf("board: failed to parse created date %q: %w", value.Value, err)
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

// Task represents a sticky note task persisted under .sticky/boards/<board>/tasks.
// The struct mirrors the YAML frontmatter (ID, Title, Status, Priority, Tags, Created, DependsOn)
// while Content holds the Markdown body and FilePath/Board* are metadata injected by repositories.
type Task struct {
	ID        string   `yaml:"id"`
	UID       string   `yaml:"uid,omitempty"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Priority  int      `yaml:"priority"`
	Tags      []string `yaml:"tags"`
	Created   Date     `yaml:"created"`
	DependsOn []string `yaml:"depends_on"`
	Content   string   `yaml:"-"`
	FilePath  string   `yaml:"-"`
	BoardID   string   `yaml:"-"`
	BoardName string   `yaml:"-"`
}

// NewTask creates a new task with default values (todo/status and default priority) and trims the title.
func NewTask(title string) (Task, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return Task{}, fmt.Errorf("board: %w", ErrInvalidTitle)
	}
	return Task{
		Title:    trimmed,
		Status:   DefaultStatus,
		Priority: DefaultPriority,
	}, nil
}

// normalizePriority enforces the allowed priority range and falls back to DefaultPriority when unspecified.
func normalizePriority(value int) (int, error) {
	if value == 0 {
		return DefaultPriority, nil
	}
	if value < 1 || value > 3 {
		return 0, fmt.Errorf("board: %w", ErrInvalidPriority)
	}
	return value, nil
}

// effectivePriority returns a priority value suitable for display, substituting the default when unset.
func effectivePriority(value int) int {
	if value == 0 {
		return DefaultPriority
	}
	return value
}
