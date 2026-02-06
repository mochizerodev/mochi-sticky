package board

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type taskFrontmatter struct {
	ID       string   `yaml:"id"`
	UID      string   `yaml:"uid,omitempty"`
	Title    string   `yaml:"title"`
	Status   string   `yaml:"status"`
	Priority int      `yaml:"priority"`
	Tags     []string `yaml:"tags"`
	Created  Date     `yaml:"created"`
	Depends  []string `yaml:"depends_on"`
}

// Parser reads and writes task files.
type Parser struct{}

// Parse converts a markdown file into a Task.
func (p *Parser) Parse(data []byte) (Task, error) {
	frontmatter, body, err := splitFrontmatter(data)
	if err != nil {
		return Task{}, err
	}

	var fm taskFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return Task{}, fmt.Errorf("board: failed to unmarshal frontmatter: %w: %v", ErrInvalidYAML, err)
	}

	task := Task{
		ID:        fm.ID,
		UID:       fm.UID,
		Title:     fm.Title,
		Status:    fm.Status,
		Priority:  fm.Priority,
		Tags:      fm.Tags,
		Created:   fm.Created,
		DependsOn: normalizeIDs(fm.Depends),
		Content:   body,
	}
	return task, nil
}

// Render converts a Task into markdown content with YAML frontmatter.
func (p *Parser) Render(task Task) ([]byte, error) {
	fm := taskFrontmatter{
		ID:       task.ID,
		UID:      task.UID,
		Title:    task.Title,
		Status:   task.Status,
		Priority: task.Priority,
		Tags:     task.Tags,
		Created:  task.Created,
		Depends:  normalizeIDs(task.DependsOn),
	}
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("board: failed to marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	if !bytes.HasSuffix(yamlBytes, []byte("\n")) {
		buf.WriteString("\n")
	}
	buf.WriteString("---\n")
	if task.Content != "" {
		buf.WriteString(task.Content)
		if !strings.HasSuffix(task.Content, "\n") {
			buf.WriteString("\n")
		}
	}
	return buf.Bytes(), nil
}

func splitFrontmatter(data []byte) (string, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	if !scanner.Scan() {
		return "", "", fmt.Errorf("board: missing frontmatter: %w", ErrInvalidFrontmatter)
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return "", "", fmt.Errorf("board: missing frontmatter: %w", ErrInvalidFrontmatter)
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
		return "", "", fmt.Errorf("board: failed to scan content: %w", err)
	}
	if inFrontmatter {
		return "", "", fmt.Errorf("board: missing frontmatter end: %w", ErrInvalidFrontmatter)
	}

	frontmatter := strings.Join(fmLines, "\n")
	body := strings.Join(bodyLines, "\n")
	return frontmatter, body, nil
}

func normalizeIDs(ids []string) []string {
	var out []string
	seen := make(map[string]struct{})
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
