# AI Development Context

## Instructions for the AI Agent
You are assisting in the development of `mochi-sticky`. Before writing code, you must load the context from:
-> MASTER_SPEC.md

## Key Constraints
1. **File-Based:** The project is a "Flat-File" system. All state is in Markdown.
2. **Standard-Library First:** Use Go standard library where possible, supplemented by the 'Charm' ecosystem (Bubble Tea).
3. **Sandboxed:** All file I/O must be localized to the `.sticky/` folder.
4. **Interface Parity:** Logic implemented for the CLI must be reusable by the TUI and MCP server.
5. **Task/Board Access Order:** When interacting with `.sticky` data, try MCP first, then CLI, and only then manual edits. If manual edits are required due to a limitation, document the limitation and a proposed improvement in "board-improvement" board.
6. **Wiki Access Order:** When interacting with `.sticky` data, try MCP first, then CLI, and only then manual edits. If manual edits are required due to a limitation, document the limitation and a proposed improvement in "wiki-improvement" board.
7. **Tests Required:** Every code change must add or update unit tests to cover the behavior change.
8. **Go Test Practices:** Tests should follow AAA (Arrange-Act-Assert), be table-driven when useful, use clear names, avoid unnecessary coupling, and prefer focused, deterministic unit tests over integration-style tests unless explicitly required.
