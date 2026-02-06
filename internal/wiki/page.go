package wiki

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type pageFrontmatter struct {
	Title   string   `yaml:"title"`
	Slug    string   `yaml:"slug"`
	Section string   `yaml:"section"`
	Order   int      `yaml:"order"`
	Tags    []string `yaml:"tags"`
	Status  string   `yaml:"status"`
}

// Page represents a wiki page loaded from disk.
type Page struct {
	Title    string
	Slug     string
	Section  string
	Order    int
	Tags     []string
	Status   string
	Content  string
	FilePath string
}

// ParsePage converts a markdown file into a Page.
func ParsePage(data []byte) (Page, error) {
	frontmatter, body, err := splitFrontmatter(data)
	if err != nil {
		return Page{}, err
	}

	var fm pageFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return Page{}, fmt.Errorf("wiki: failed to unmarshal frontmatter: %w: %v", ErrInvalidYAML, err)
	}

	return Page{
		Title:   fm.Title,
		Slug:    fm.Slug,
		Section: fm.Section,
		Order:   fm.Order,
		Tags:    fm.Tags,
		Status:  fm.Status,
		Content: body,
	}, nil
}

// RenderPage converts a Page into markdown content with YAML frontmatter.
func RenderPage(page Page) ([]byte, error) {
	fm := pageFrontmatter{
		Title:   page.Title,
		Slug:    page.Slug,
		Section: page.Section,
		Order:   page.Order,
		Tags:    page.Tags,
		Status:  page.Status,
	}
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("wiki: failed to marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlBytes)
	if !bytes.HasSuffix(yamlBytes, []byte("\n")) {
		buf.WriteString("\n")
	}
	buf.WriteString("---\n")
	if page.Content != "" {
		buf.WriteString(page.Content)
		if !strings.HasSuffix(page.Content, "\n") {
			buf.WriteString("\n")
		}
	}
	return buf.Bytes(), nil
}

// LoadPage reads a markdown file from disk and parses it into a Page.
func LoadPage(path string) (Page, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Page{}, fmt.Errorf("wiki: failed to read page %s: %w", path, err)
	}
	page, err := ParsePage(data)
	if err != nil {
		return Page{}, err
	}
	page.FilePath = path
	return page, nil
}

// SavePage renders and writes a Page to disk.
func SavePage(path string, page Page) error {
	data, err := RenderPage(page)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("wiki: failed to create page dir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("wiki: failed to write page %s: %w", path, err)
	}
	return nil
}

func splitFrontmatter(data []byte) (string, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	if !scanner.Scan() {
		return "", "", fmt.Errorf("wiki: missing frontmatter: %w", ErrInvalidFrontmatter)
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return "", "", fmt.Errorf("wiki: missing frontmatter: %w", ErrInvalidFrontmatter)
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
		return "", "", fmt.Errorf("wiki: failed to scan content: %w", err)
	}
	if inFrontmatter {
		return "", "", fmt.Errorf("wiki: missing frontmatter end: %w", ErrInvalidFrontmatter)
	}

	frontmatter := strings.Join(fmLines, "\n")
	body := strings.Join(bodyLines, "\n")
	return frontmatter, body, nil
}
