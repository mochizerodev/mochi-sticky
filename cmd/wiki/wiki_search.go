package wiki

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/shared"

	"github.com/spf13/cobra"
)

var wikiSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search wiki pages",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		storageRoot, err := cli.ResolveStorageRoot(workingDir, false)
		if err != nil {
			return err
		}
		root := wikiRoot(storageRoot)
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No matches found.")
				return err
			}
			return err
		}

		statusFilter, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}
		includeTemplates, err := cmd.Flags().GetBool("include-templates")
		if err != nil {
			return err
		}
		templatePaths, err := cli.ResolveTemplatePaths(workingDir, storageRoot)
		if err != nil {
			return err
		}
		templatesRoot := ""
		if includeTemplates && strings.TrimSpace(templatePaths.Wiki) != "" && !shared.IsSubpath(root, templatePaths.Wiki) {
			if info, err := os.Stat(templatePaths.Wiki); err == nil && info.IsDir() {
				templatesRoot = templatePaths.Wiki
			}
		}

		if err := searchWithRipgrep(cmd, root, templatesRoot, query, statusFilter, includeTemplates); err == nil {
			return nil
		} else if !errors.Is(err, errRipgrepUnavailable) {
			return err
		}

		return searchWithGo(cmd, root, templatesRoot, query, statusFilter, includeTemplates)
	},
}

func init() {
	wikiCmd.AddCommand(wikiSearchCmd)
	wikiSearchCmd.Flags().String("status", "", "Filter by status (draft|published|archived)")
	wikiSearchCmd.Flags().Bool("include-templates", false, "Include template pages in search results")
}

var errRipgrepUnavailable = errors.New("ripgrep unavailable")

func searchWithRipgrep(
	cmd *cobra.Command,
	root, templatesRoot, query, statusFilter string,
	includeTemplates bool,
) error {
	if _, err := exec.LookPath("rg"); err != nil {
		return errRipgrepUnavailable
	}

	if strings.TrimSpace(statusFilter) != "" {
		return errRipgrepUnavailable
	}

	searchRoot := func(searchBase string, includeTemplates bool) ([]string, error) {
		rgArgs := []string{"-F", "-i", "--with-filename", "--line-number", "--no-heading", "--color=never", query, searchBase}
		if !includeTemplates {
			rgArgs = append(rgArgs, "--glob", "!**/templates/**")
		}
		rgCmd := exec.Command("rg", rgArgs...)
		output, err := rgCmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return nil, nil
			}
			return nil, fmt.Errorf("rg failed: %w", err)
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			return nil, nil
		}
		return lines, nil
	}

	lines, err := searchRoot(root, includeTemplates)
	if err != nil {
		return err
	}
	if includeTemplates && templatesRoot != "" {
		extra, err := searchRoot(templatesRoot, true)
		if err != nil {
			return err
		}
		lines = append(lines, extra...)
	}
	if len(lines) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No matches found.")
		return err
	}
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		searchBase := root
		if templatesRoot != "" && shared.IsSubpath(templatesRoot, parts[0]) {
			searchBase = templatesRoot
		}
		slug := slugFromPath(searchBase, parts[0])
		if slug == "" {
			slug = parts[0]
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s:%s:%s\n", slug, parts[1], parts[2]); err != nil {
			return err
		}
	}
	return nil
}

func searchWithGo(cmd *cobra.Command, root, templatesRoot, query, statusFilter string, includeTemplates bool) error {
	found := false
	searchRoot := func(searchBase string, includeTemplates bool) error {
		return filepath.WalkDir(searchBase, func(path string, d fs.DirEntry, err error) error {
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
			statusValue := ""
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				if strings.TrimSpace(line) == "---" {
					if !inFrontmatter {
						inFrontmatter = true
						continue
					}
					if inFrontmatter {
						if strings.TrimSpace(statusFilter) != "" {
							if !strings.EqualFold(strings.TrimSpace(statusValue), strings.TrimSpace(statusFilter)) {
								return nil
							}
						}
						inFrontmatter = false
						continue
					}
				}
				if inFrontmatter && strings.HasPrefix(strings.TrimSpace(line), "status:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						statusValue = strings.TrimSpace(parts[1])
						continue
					}
				}
				if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
					slug := slugFromPath(searchBase, path)
					if slug == "" {
						slug = path
					}
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s:%d:%s\n", slug, lineNum, line); err != nil {
						return err
					}
					found = true
				}
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		})
	}

	if err := searchRoot(root, includeTemplates); err != nil {
		return err
	}
	if includeTemplates && templatesRoot != "" {
		if err := searchRoot(templatesRoot, true); err != nil {
			return err
		}
	}
	if !found {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No matches found.")
		return err
	}
	return nil
}
