package adr

import "errors"

var (
	// ErrADRNotFound indicates an ADR could not be located on disk.
	ErrADRNotFound = errors.New("adr not found")
	// ErrInvalidConfig indicates the ADR config file is invalid.
	ErrInvalidConfig = errors.New("invalid config")
	// ErrInvalidFrontmatter indicates the ADR markdown frontmatter is missing or malformed.
	ErrInvalidFrontmatter = errors.New("invalid frontmatter")
	// ErrInvalidID indicates the ADR identifier could not be parsed.
	ErrInvalidID = errors.New("invalid id")
	// ErrInvalidTitle indicates the ADR title is missing or invalid.
	ErrInvalidTitle = errors.New("invalid title")
	// ErrInvalidYAML indicates the ADR frontmatter YAML could not be parsed.
	ErrInvalidYAML = errors.New("invalid yaml")
)
