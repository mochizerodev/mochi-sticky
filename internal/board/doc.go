// Package board contains the domain logic for boards and sticky-note tasks.
// It exposes repositories, parsers, and helpers for reading/writing task Markdown files
// under `.sticky/boards/<board-id>/tasks/`. The package defines the canonical Task schema,
// filters tasks by status/tags, manages dependencies, and provides utilities used by the CLI,
// the TUI, and MCP handlers.
package board
