package wiki

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/shared"
)

// SearchResult captures a single search hit.
type SearchResult struct {
	Slug    string `json:"slug"`
	Line    int    `json:"line"`
	Snippet string `json:"snippet"`
}

// SearchOptions controls wiki search behavior.
type SearchOptions struct {
	Query            string
	Status           string
	IncludeTemplates bool
	CaseInsensitive  bool
	TemplatesRoot    string
}

// SearchPages scans wiki pages for the query and returns matching lines.
func SearchPages(root string, opts SearchOptions) ([]SearchResult, error) {
	return SearchPagesContext(context.Background(), root, opts)
}

// SearchPagesContext scans wiki pages for the query and returns matching lines, honoring ctx cancellation.
func SearchPagesContext(ctx context.Context, root string, opts SearchOptions) ([]SearchResult, error) {
	if strings.TrimSpace(opts.Query) == "" {
		return nil, fmt.Errorf("wiki: search query is required")
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("wiki: failed to stat wiki root %s: %w", root, err)
	}

	query := opts.Query
	if opts.CaseInsensitive {
		query = strings.ToLower(query)
	}
	statusFilter := strings.ToLower(strings.TrimSpace(opts.Status))

	results := make([]SearchResult, 0)
	searchRoot := func(searchBase string, includeTemplates bool) error {
		return filepath.WalkDir(searchBase, func(path string, d fs.DirEntry, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if err != nil {
				return err
			}
			if d.IsDir() {
				if !includeTemplates && filepath.Base(path) == "templates" {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) != ".md" {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close()
			}()

			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			lineNum := 0
			inFrontmatter := false
			pageStatus := ""
			allowed := true
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				lineNum++
				line := scanner.Text()
				if strings.TrimSpace(line) == "---" {
					if !inFrontmatter {
						inFrontmatter = true
						continue
					}
					if inFrontmatter {
						if statusFilter != "" {
							if strings.ToLower(strings.TrimSpace(pageStatus)) != statusFilter {
								allowed = false
							}
						}
						inFrontmatter = false
						continue
					}
				}
				if inFrontmatter && strings.HasPrefix(strings.TrimSpace(line), "status:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						pageStatus = strings.TrimSpace(parts[1])
						continue
					}
				}
				if !allowed {
					continue
				}
				haystack := line
				if opts.CaseInsensitive {
					haystack = strings.ToLower(line)
				}
				if strings.Contains(haystack, query) {
					slug := SlugFromPath(searchBase, path)
					if slug == "" {
						slug = path
					}
					results = append(results, SearchResult{
						Slug:    slug,
						Line:    lineNum,
						Snippet: line,
					})
				}
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		})
	}

	if err := searchRoot(root, opts.IncludeTemplates); err != nil {
		return nil, fmt.Errorf("wiki: failed to search pages: %w", err)
	}
	if opts.IncludeTemplates &&
		strings.TrimSpace(opts.TemplatesRoot) != "" &&
		!shared.IsSubpath(root, opts.TemplatesRoot) {
		if info, err := os.Stat(opts.TemplatesRoot); err == nil {
			if info.IsDir() {
				if err := searchRoot(opts.TemplatesRoot, true); err != nil {
					return nil, fmt.Errorf("wiki: failed to search templates: %w", err)
				}
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("wiki: failed to stat templates root %s: %w", opts.TemplatesRoot, err)
		}
	}
	return results, nil
}
