# GitHub Copilot Instructions

## Primary Source of Truth
CRITICAL: Always refer to @workspace /MASTER_SPEC.md for architectural definitions and directory structures.
Strictly adhere to the coding standards defined in MASTER_SPEC.md Section 6. If a suggested snippet violates these rules correct it before presenting

## Coding Standards (Go)
- **Frameworks:** Strictly use `cobra` for CLI and `bubbletea` for TUI.
- **Styling:** Use `lipgloss` for all terminal UI formatting.
- **Error Handling:** Avoid panics. Use `fmt.Errorf("context: %w", err)`.
- **Concurrency:** Prefer channels for TUI state updates from the file system.
- **Naming:** Follow idiomatic Go conventions (PascalCase for exported, camelCase for internal).

## Prohibited Patterns
- DO NOT suggest using SQLite or any external database.
- DO NOT use raw ANSI codes for colors (use lipgloss).