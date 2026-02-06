// Package tui implements the interactive Bubble Tea terminal interface that
// mirrors the mochi-sticky board/wiki/adr experience in a TUI.
// The package wires the board repository into a `Model` that tracks layout,
// pagination, view state (kanban, task detail, wiki widgets, confirmations), and
// editing inputs. `Run`/`RunWithEditor` instantiate the model, open the alternate
// screen buffer, and enter the Bubble Tea event loop, while `view.go` renders the
// themed layout with lipgloss styles and helpers.
package tui
