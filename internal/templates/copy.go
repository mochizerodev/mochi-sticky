package templates

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func copyDirIfEmpty(sourceDir, targetDir string) error {
	if sourceDir == "" || targetDir == "" {
		return nil
	}
	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("templates: failed to stat source %s: %w", sourceDir, err)
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("templates: expected directory for source %s", sourceDir)
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return fmt.Errorf("templates: failed to create target %s: %w", targetDir, err)
			}
		} else {
			return fmt.Errorf("templates: failed to read target %s: %w", targetDir, err)
		}
	} else if len(entries) > 0 {
		return nil
	}

	return filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(targetDir, rel)
		if d.IsDir() {
			if rel == "." {
				return nil
			}
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return fmt.Errorf("templates: failed to create directory %s: %w", destPath, err)
			}
			return nil
		}
		if err := copyFile(path, destPath); err != nil {
			return err
		}
		return nil
	})
}

func copyFile(sourcePath, targetPath string) error {
	input, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("templates: failed to open source %s: %w", sourcePath, err)
	}
	defer func() {
		_ = input.Close()
	}()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("templates: failed to create target directory %s: %w", filepath.Dir(targetPath), err)
	}
	output, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("templates: failed to create target %s: %w", targetPath, err)
	}
	defer func() {
		_ = output.Close()
	}()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("templates: failed to copy %s to %s: %w", sourcePath, targetPath, err)
	}
	return nil
}

func copyEmbeddedDirIfEmpty(sourceDir, targetDir string) error {
	if sourceDir == "" || targetDir == "" {
		return nil
	}
	if !strings.HasPrefix(sourceDir, "assets/") {
		return fmt.Errorf("templates: invalid embedded source %s", sourceDir)
	}

	_, err := fs.ReadDir(embeddedFS, sourceDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("templates: failed to read embedded source %s: %w", sourceDir, err)
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return fmt.Errorf("templates: failed to create target %s: %w", targetDir, err)
			}
		} else {
			return fmt.Errorf("templates: failed to read target %s: %w", targetDir, err)
		}
	} else if len(entries) > 0 {
		return nil
	}

	return fs.WalkDir(embeddedFS, sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		destPath := filepath.Join(targetDir, filepath.FromSlash(rel))
		if d.IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return fmt.Errorf("templates: failed to create directory %s: %w", destPath, err)
			}
			return nil
		}
		return copyEmbeddedFile(path, destPath)
	})
}

func copyEmbeddedFileIfMissing(sourcePath, targetPath string) error {
	if sourcePath == "" || targetPath == "" {
		return nil
	}
	if !strings.HasPrefix(sourcePath, "assets/") {
		return fmt.Errorf("templates: invalid embedded source %s", sourcePath)
	}
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("templates: failed to stat target %s: %w", targetPath, err)
	}
	return copyEmbeddedFile(sourcePath, targetPath)
}

func copyEmbeddedFile(sourcePath, targetPath string) error {
	input, err := embeddedFS.Open(sourcePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("templates: failed to open embedded source %s: %w", sourcePath, err)
	}
	defer func() {
		_ = input.Close()
	}()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("templates: failed to create target directory %s: %w", filepath.Dir(targetPath), err)
	}
	output, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("templates: failed to create target %s: %w", targetPath, err)
	}
	defer func() {
		_ = output.Close()
	}()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("templates: failed to copy embedded %s to %s: %w", sourcePath, targetPath, err)
	}
	return nil
}
