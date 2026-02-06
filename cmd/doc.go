// Package cmd configures the `mochi-sticky` Cobra CLI, wiring every task, board,
// ADR, wiki, MCP, and helper command into the single entry point used by `main.go`.
// `root.go` defines the base command along with the shared `--storage` flag
// that overrides config/env defaults, while each subcommand (add, list, status,
// wiki-*, MCP, etc.) lives in its own file and attaches flags, validation, and
// behavior under the root command tree.
package cmd
