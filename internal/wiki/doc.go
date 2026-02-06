// Package wiki encapsulates the workspace wiki domain, from reading and writing
// markdown pages with YAML frontmatter to managing indexes, manifests, exports,
// and search/filter helpers. It defines the data models for pages, sections, and
// navigation nodes, offers validation and linting rules, and exposes functions to
// generate filtered lists, render PDF exports, and build the in-repo index used
// by both the CLI/TUI and the MCP layer.
package wiki
