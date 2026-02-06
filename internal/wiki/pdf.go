package wiki

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PDFOptions configures PDF export via pandoc.
type PDFOptions struct {
	Output   string
	Title    string
	Author   string
	Template string
	BaseDir  string
	TempDir  string
}

// WritePDF writes a PDF export using pandoc.
func WritePDF(markdown []byte, opts PDFOptions) (string, error) {
	return WritePDFContext(context.Background(), markdown, opts)
}

// WritePDFContext writes a PDF export using pandoc and honors ctx cancellation.
func WritePDFContext(ctx context.Context, markdown []byte, opts PDFOptions) (string, error) {
	output := strings.TrimSpace(opts.Output)
	if output == "" {
		return "", fmt.Errorf("wiki: output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", fmt.Errorf("wiki: failed to create export dir %s: %w", filepath.Dir(output), err)
	}
	if _, err := exec.LookPath("pandoc"); err != nil {
		return "", fmt.Errorf("wiki: pandoc is required for pdf export")
	}

	tempDir := strings.TrimSpace(opts.TempDir)
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	tmp, err := os.CreateTemp(tempDir, "export-*.md")
	if err != nil {
		return "", fmt.Errorf("wiki: failed to create temp export: %w", err)
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	if _, err := tmp.Write(markdown); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("wiki: failed to write temp export: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("wiki: failed to close temp export: %w", err)
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	args := []string{tmp.Name(), "-o", output, "--toc"}
	if strings.TrimSpace(opts.Title) != "" {
		args = append(args, "--metadata", "title="+strings.TrimSpace(opts.Title))
	}
	if strings.TrimSpace(opts.Author) != "" {
		args = append(args, "--metadata", "author="+strings.TrimSpace(opts.Author))
	}
	template := strings.TrimSpace(opts.Template)
	if template == "" && strings.TrimSpace(opts.BaseDir) != "" {
		defaultTemplate := filepath.Join(opts.BaseDir, "docs", "wiki_pdf_template.tex")
		if _, err := os.Stat(defaultTemplate); err == nil {
			template = defaultTemplate
		}
	}
	if template != "" {
		args = append(args, "--template", template)
	}

	pandocCmd := exec.CommandContext(ctx, "pandoc", args...)
	if out, err := pandocCmd.CombinedOutput(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("wiki: pandoc failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return output, nil
}
