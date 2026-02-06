package adr

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type adrFrontmatter struct {
	ID           int      `yaml:"id"`
	UID          string   `yaml:"uid,omitempty"`
	Title        string   `yaml:"title"`
	Status       string   `yaml:"status"`
	Date         Date     `yaml:"date"`
	Tags         []string `yaml:"tags,omitempty"`
	Supersedes   []int    `yaml:"supersedes,omitempty"`
	SupersededBy int      `yaml:"superseded_by,omitempty"`
	Links        []string `yaml:"links,omitempty"`
}

// ADR represents an Architecture Decision Record persisted under `.sticky/adrs/*.md`.
type ADR struct {
	ID           int
	UID          string
	Title        string
	Status       string
	Date         Date
	Tags         []string
	Supersedes   []int
	SupersededBy int
	Links        []string
	Content      string
	FilePath     string
}

// FormatID renders an ADR numeric ID as a zero-padded 4-digit string (e.g., 0001).
func FormatID(id int) string {
	return fmt.Sprintf("%04d", id)
}

// ParseID parses an ADR identifier from user input (e.g., "1", "0001", "ADR-0001").
func ParseID(input string) (int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, fmt.Errorf("adr: %w", ErrInvalidID)
	}
	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "ADR-") {
		trimmed = strings.TrimSpace(trimmed[4:])
	}
	trimmed = strings.TrimLeft(trimmed, "0")
	if trimmed == "" {
		trimmed = "0"
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("adr: %w", ErrInvalidID)
	}
	return value, nil
}

// Slugify converts a string into a file-safe slug.
func Slugify(value string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			prevDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// DefaultContent returns a baseline ADR body template with required sections.
func DefaultContent() string {
	return strings.TrimLeft(`
## Context

What is the problem we’re trying to solve?

## Decision

What did we decide and why?

## Consequences

What becomes easier/harder because of this decision?
`, "\n")
}

// ParseADR converts a markdown file into an ADR.
func ParseADR(data []byte) (ADR, error) {
	frontmatter, body, err := splitFrontmatter(data)
	if err != nil {
		return ADR{}, err
	}

	var fm adrFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return ADR{}, fmt.Errorf("adr: failed to unmarshal frontmatter: %w: %v", ErrInvalidYAML, err)
	}
	return ADR{
		ID:           fm.ID,
		UID:          fm.UID,
		Title:        fm.Title,
		Status:       fm.Status,
		Date:         fm.Date,
		Tags:         fm.Tags,
		Supersedes:   fm.Supersedes,
		SupersededBy: fm.SupersededBy,
		Links:        fm.Links,
		Content:      body,
	}, nil
}

// RenderADR converts an ADR into markdown content with YAML frontmatter.
func RenderADR(adr ADR) ([]byte, error) {
	if strings.TrimSpace(adr.Title) == "" {
		return nil, fmt.Errorf("adr: %w", ErrInvalidTitle)
	}
	if adr.ID <= 0 {
		return nil, fmt.Errorf("adr: %w", ErrInvalidID)
	}

	fm := adrFrontmatter{
		ID:           adr.ID,
		UID:          adr.UID,
		Title:        adr.Title,
		Status:       adr.Status,
		Date:         adr.Date,
		Tags:         adr.Tags,
		Supersedes:   adr.Supersedes,
		SupersededBy: adr.SupersededBy,
		Links:        adr.Links,
	}
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("adr: failed to marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	if !bytes.HasSuffix(yamlBytes, []byte("\n")) {
		buf.WriteString("\n")
	}
	buf.WriteString("---\n")
	if adr.Content != "" {
		buf.WriteString(adr.Content)
		if !strings.HasSuffix(adr.Content, "\n") {
			buf.WriteString("\n")
		}
	}
	return buf.Bytes(), nil
}

// LoadADR reads a markdown file from disk and parses it into an ADR.
func LoadADR(path string) (ADR, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ADR{}, fmt.Errorf("adr: failed to read adr %s: %w", path, err)
	}
	adr, err := ParseADR(data)
	if err != nil {
		return ADR{}, err
	}
	adr.FilePath = path

	if idFromName, ok := parseIDFromFilename(filepath.Base(path)); ok {
		if adr.ID == 0 {
			adr.ID = idFromName
		} else if adr.ID != idFromName {
			return ADR{}, fmt.Errorf("adr: id mismatch for %s (frontmatter=%d filename=%d)", path, adr.ID, idFromName)
		}
	}

	return adr, nil
}

// SaveADR renders and writes an ADR to disk.
func SaveADR(path string, adr ADR) error {
	data, err := RenderADR(adr)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("adr: failed to create adr dir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("adr: failed to write adr %s: %w", path, err)
	}
	return nil
}

func splitFrontmatter(data []byte) (string, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	if !scanner.Scan() {
		return "", "", fmt.Errorf("adr: missing frontmatter: %w", ErrInvalidFrontmatter)
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return "", "", fmt.Errorf("adr: missing frontmatter: %w", ErrInvalidFrontmatter)
	}

	var fmLines []string
	var bodyLines []string
	inFrontmatter := true

	for scanner.Scan() {
		line := scanner.Text()
		if inFrontmatter && strings.TrimSpace(line) == "---" {
			inFrontmatter = false
			continue
		}
		if inFrontmatter {
			fmLines = append(fmLines, line)
			continue
		}
		bodyLines = append(bodyLines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("adr: failed to scan content: %w", err)
	}
	if inFrontmatter {
		return "", "", fmt.Errorf("adr: missing frontmatter end: %w", ErrInvalidFrontmatter)
	}

	frontmatter := strings.Join(fmLines, "\n")
	body := strings.Join(bodyLines, "\n")
	return frontmatter, body, nil
}

func parseIDFromFilename(name string) (int, bool) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	base = strings.TrimSpace(base)
	if base == "" {
		return 0, false
	}
	var digits strings.Builder
	for _, r := range base {
		if r < '0' || r > '9' {
			break
		}
		digits.WriteRune(r)
	}
	if digits.Len() == 0 {
		return 0, false
	}
	value, err := strconv.Atoi(digits.String())
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}
