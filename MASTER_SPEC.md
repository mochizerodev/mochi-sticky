# MASTER SPEC: mochi-sticky 🍡

## 1. Vision & Branding
`mochi-sticky` is a developer-first project management tool. It treats Markdown files as "sticky notes" on a Kanban board.
- **Brand Identity:** Soft, clean, and efficient (inspired by Mochi).
- **Core Value:** Data is code. Everything lives in the Git repo.

## 2. Technical Stack
- **Language:** Go (Golang)
- **CLI:** spf13/cobra
- **TUI:** charmbracelet/bubbletea & charmbracelet/lipgloss
- **Persistence:** Local Markdown files with YAML Frontmatter.
- **Agent Interface:** Model Context Protocol (MCP) via JSON-RPC.

## 3. Data Schema (The "Sticky")
All tasks reside in `.sticky/boards/<board-id>/tasks/*.md` (legacy `.sticky/tasks` is migrated on init).
```yaml
---
id: "unique-uuid"
title: "Task Title"
status: "<column key from .sticky/config.yaml>"
priority: 1 | 2 | 3
tags: [tag1, tag2]
created: YYYY-MM-DD
---
# Description
Markdown content goes here.
```

Statuses are config-driven. The `status` field must match a column `key` in `.sticky/config.yaml`.

## 4. Architecture
- A. Domain + Storage
    - `internal/board`: Task + board domain models, filters, repositories.
    - `internal/storage`: Storage root/config hydration and global settings.
    - `internal/shared`: Cross-cutting helpers (path validation).
- B. User Interfaces
    - CLI: Rapid commands (add, list, move).
    - TUI: Multi-column Kanban board with keyboard navigation (h,j,k,l).
    - MCP: Standard I/O server allowing agents to manipulate the board.

## 5. Directory Structure

```
mochi-sticky/
├── .sticky/         # User data (boards/wiki/config)
├── cmd/             # Cobra command definitions (thin wiring)
├── internal/        # Domain + TUI + MCP logic
│   ├── adr/         # ADR domain helpers/templates
│   ├── board/       # Board/task domain + repositories
│   ├── mcp/         # MCP server + adapters
│   ├── shared/      # Cross-cutting helpers
│   ├── storage/     # Storage root/config + settings
│   ├── tui/         # Terminal UI
│   └── wiki/        # Wiki domain logic
└── MASTER_SPEC.md   # This file
```


## 6. Coding Standards (The Mochi Way)

### A. Error Handling
- **No Panics:** Never use `panic()` or `os.Exit()` inside `internal/`.
- **Wrapping:** Always wrap errors with context: `fmt.Errorf("board: failed to parse task %s: %w", id, err)`.
- **Sentinels:** Define custom error types in `internal/board/errors.go` for common failures (e.g., `ErrTaskNotFound`).

### B. Concurrency & State
- **Thread Safety:** The `Repository` must be thread-safe (use `sync.RWMutex`).
- **TUI Messages:** In Bubble Tea, never update the model directly from a goroutine; always send a `tea.Cmd`.

### C. Style & Documentation
- **Minimalist:** Keep functions small and focused on a single task.
- **Self-Documenting:** Use descriptive variable names (`taskFilePath` instead of `tp`).
- **Comments:** Every exported function MUST have a comment starting with the function name.
- **Interfaces:** Accept interfaces, return structs.

### D. Testing
- **Table-Driven Tests:** Use Go’s table-driven test pattern for all logic in `internal/board`.
- **Mocking:** Create mock file systems for testing the `Repository` without hitting the actual disk.
- **AAA Pattern:** Structure tests as Arrange-Act-Assert with clear sectioning and minimal setup per case.
- **Coverage Expectation:** Every code change must add or update unit tests that cover the behavior change.
- **Best Practices:** Prefer focused, deterministic unit tests; use descriptive test names; avoid over-mocking and unnecessary coupling; use helper functions for shared setup; and keep assertions specific to the behavior under test.
- **Integration Tests:** Add or update integration tests when changes affect CLI wiring, end-to-end workflows, or cross-package behavior (e.g., `cmd/` commands, storage initialization, exports, or multi-step flows).
- **Integration Scope:** Integration tests should run the CLI against a temp storage root and assert user-visible behavior (output text, files created/modified, and side effects). These do not replace unit tests.
- **Unit Test Focus:** Unit tests should validate domain logic, parsing/serialization, filtering/sorting, and error cases at the package level (primarily under `internal/`). When adding integration coverage, still add unit tests for the underlying logic and edge cases.

### E. Output & Logging Discipline (CRITICAL)
- **Stdout is Forbidden:** Never use `fmt.Println` or `log.Println` for debugging.
    - **In TUI:** It corrupts the UI buffer.
    - **In MCP:** It breaks the JSON-RPC pipe (Agent crashes).
- **Logging Strategy:**
    - Use `tea.LogToFile("debug.log", "prefix")` for TUI debugging.
    - For the Core Logic, accept a `Logger` interface so the caller can decide where logs go (File vs Stderr vs Discard).
- **Stderr Usage:** Reserved exclusively for fatal application errors that occur *before* the TUI/MCP starts.

### F. File System Hygiene (Cross-Platform)
- **Path Separators:** Never hardcode strings like `"tasks/" + id`. Always use `filepath.Join("tasks", id)`.
- **Windows Compatibility:** The AI must handle Windows newline characters (`\r\n`) when parsing Markdown headers.
- **Git Ignore:** Ensure the code automatically adds `.sticky/debug.log` to a `.gitignore` file inside `.sticky/` so logs aren't committed.

### G. License Hygiene (CRITICAL)
- **No Unlicensed Copy/Paste:** Don’t paste in code you don’t have rights to (license compatibility). When in doubt, re-implement from first principles or use permissively licensed sources with proper attribution.
