package wiki

import "errors"

// ErrInvalidFrontmatter indicates the frontmatter section is missing or malformed.
var ErrInvalidFrontmatter = errors.New("invalid frontmatter")

// ErrInvalidYAML indicates the frontmatter YAML could not be unmarshaled.
var ErrInvalidYAML = errors.New("invalid yaml")

// ErrIndexNotFound indicates the index file is missing.
var ErrIndexNotFound = errors.New("index not found")

// ErrDuplicateSlug indicates multiple pages share the same slug.
var ErrDuplicateSlug = errors.New("duplicate slug")

// ErrPageNotFound indicates a requested page does not exist.
var ErrPageNotFound = errors.New("page not found")

// ErrSectionNotFound indicates a requested section does not exist.
var ErrSectionNotFound = errors.New("section not found")
