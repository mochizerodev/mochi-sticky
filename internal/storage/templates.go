package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplatesConfig captures template directory overrides.
type TemplatesConfig struct {
	Root    string `yaml:"root"`
	ADR     string `yaml:"adr"`
	Task    string `yaml:"task"`
	Board   string `yaml:"board"`
	Wiki    string `yaml:"wiki"`
	WikiPDF string `yaml:"wiki_pdf"`
}

// TemplatePaths contains resolved template locations.
type TemplatePaths struct {
	Root    string
	ADR     string
	Task    string
	Board   string
	Wiki    string
	WikiPDF string
}

// ResolveTemplates determines template locations from config, defaults, and legacy paths.
func ResolveTemplates(workingDir, storageRoot string, cfg Config) (TemplatePaths, error) {
	if strings.TrimSpace(workingDir) == "" {
		return TemplatePaths{}, fmt.Errorf("storage: working directory is required")
	}
	if strings.TrimSpace(storageRoot) == "" {
		return TemplatePaths{}, fmt.Errorf("storage: storage root is required")
	}

	root := strings.TrimSpace(cfg.Templates.Root)
	if root == "" {
		root = filepath.Join(storageRoot, "templates")
	}
	rootPath, err := resolveTemplateDir(workingDir, root, true)
	if err != nil {
		return TemplatePaths{}, err
	}

	adrPath, err := resolveTemplateDir(workingDir, fallback(cfg.Templates.ADR, filepath.Join(rootPath, "adr")), true)
	if err != nil {
		return TemplatePaths{}, err
	}
	wikiPath, err := resolveTemplateDir(workingDir, fallback(cfg.Templates.Wiki, filepath.Join(rootPath, "wiki")), true)
	if err != nil {
		return TemplatePaths{}, err
	}
	taskPath, err := resolveTemplateDir(workingDir, fallback(cfg.Templates.Task, filepath.Join(rootPath, "task")), true)
	if err != nil {
		return TemplatePaths{}, err
	}
	boardPath, err := resolveTemplateDir(workingDir, fallback(cfg.Templates.Board, filepath.Join(rootPath, "board")), true)
	if err != nil {
		return TemplatePaths{}, err
	}

	pdfTemplate := strings.TrimSpace(cfg.Templates.WikiPDF)
	if pdfTemplate == "" {
		pdfTemplate = strings.TrimSpace(cfg.PDFTemplate)
	}
	if pdfTemplate == "" {
		pdfTemplate = filepath.Join(rootPath, "wiki", "wiki_pdf_template.tex")
	}
	pdfPath, err := resolveTemplateFile(workingDir, pdfTemplate, true)
	if err != nil {
		return TemplatePaths{}, err
	}

	return TemplatePaths{
		Root:    rootPath,
		ADR:     adrPath,
		Task:    taskPath,
		Board:   boardPath,
		Wiki:    wikiPath,
		WikiPDF: pdfPath,
	}, nil
}

func resolveTemplateDir(workingDir, value string, allowMissing bool) (string, error) {
	return resolveTemplatePath(workingDir, value, allowMissing, true)
}

func resolveTemplateFile(workingDir, value string, allowMissing bool) (string, error) {
	return resolveTemplatePath(workingDir, value, allowMissing, false)
}

func resolveTemplatePath(workingDir, value string, allowMissing, expectDir bool) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("storage: template path is required")
	}
	if !filepath.IsAbs(trimmed) {
		trimmed = filepath.Join(workingDir, trimmed)
	}
	absPath, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("storage: failed to resolve template path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) && allowMissing {
			return absPath, nil
		}
		return "", fmt.Errorf("storage: failed to stat template path %s: %w", absPath, err)
	}
	if expectDir && !info.IsDir() {
		return "", fmt.Errorf("storage: template path is not a directory: %s", absPath)
	}
	if !expectDir && info.IsDir() {
		return "", fmt.Errorf("storage: template path is a directory: %s", absPath)
	}
	return absPath, nil
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}
