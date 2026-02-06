---
title: Architecture
slug: development/architecture
section: Development
order: 40
tags:
    - architecture
status: published
---
# Architecture

mochi-sticky follows a **Decoupled Layered Architecture** where core domain logic is independent of the CLI/TUI and external integrations (MCP). This document serves as both a guide and a context map for the system.

## Architectural Philosophy

**Key Principles**

1. **Single Source of Truth** — The file system (`.sticky/`) is the database
2. **Statelessness** — Internal state can be rebuilt entirely from Markdown files
3. **Concurrency-Safe** — Multiple interfaces can operate simultaneously

## System Components

### 1. Domain + Storage (`internal/board`, `internal/storage`, `internal/shared`)

The heart of mochi-sticky, handling all business logic:

**Board Domain (`internal/board/`)**
- Task and board models with YAML frontmatter
- Repository pattern for file persistence
- Filtering, sorting, and dependency management
- Archive and restore operations
- Thread-safe operations with `sync.RWMutex`

**Storage (`internal/storage/`)**
- Storage root resolution (CLI flag → env → config → default)
- Global settings management (editor preferences)
- Template directory hydration
- Configuration validation

**Shared (`internal/shared/`)**
- Path validation and sanitization
- Cross-cutting error types
- Common utilities

### 2. User Interfaces

**CLI (`cmd/`)**
- Thin command wrappers using spf13/cobra
- Direct calls to domain layer
- Output formatting and error handling
- Command composition and flags

**TUI (`internal/tui/`)**
- Built with Bubble Tea (The Elm Architecture)
- Model: Current view state (tasks, columns, focus)
- Update: Message processing (keypresses, file events)
- View: Rendering with lipgloss styling
- Multi-panel layout: Kanban board + detail view

### 3. MCP Server (`internal/mcp/`)

JSON-RPC 2.0 server for AI agent integration:

**Tools**
- Task operations: create, update, move, archive, delete
- Board management: list, create, rename, switch
- Wiki operations: read, write, search, export
- Configuration queries

**Resources**
- Read-only access to task markdown
- Board configuration and descriptions
- Board registry data
- Wiki content

**Safety**
- Destructive operations require `force: true`
- Board scoping with explicit `board_id`
- Standard error responses

### 4. Wiki & ADR (`internal/wiki/`, `internal/adr/`)

**Wiki**
- Markdown file I/O with YAML frontmatter
- Full-text search with filtering
- Section-based organization
- Index generation
- Multi-format export (Markdown, PDF)
- Template system

**ADR**
- Architecture Decision Record templates
- Status tracking (proposed → accepted → deprecated)
- Linking between ADRs
- Date-based filtering

## Architecture Diagrams

See `development/architecture-diagrams` for Mermaid diagrams (system context, component view, and data flows).

## Data Flow

### Write Operations

```
User Input → Validation → Domain Logic → File Update → Refresh
     ↓           ↓             ↓              ↓           ↓
  TUI/CLI    Business      Repository    Markdown    Watchers
             Rules         Layer         Files       Notify
```

1. **Input**: User action (keypress in TUI, CLI command, or MCP tool call)
2. **Validation**: Domain layer validates the operation
3. **Logic**: Business rules applied (e.g., status transitions)
4. **Persistence**: Repository updates the `.md` file with new YAML frontmatter
5. **Sync**: File watchers detect changes and notify TUI for refresh

### Read Operations

```
Query → Repository → Parser → Filter/Sort → Format → Display
  ↓         ↓          ↓          ↓           ↓        ↓
User     File I/O   Markdown   Domain     Presenter  Output
         Layer      Parser     Logic      Layer
```

## Technical Specifications

### Storage Root & Config Hydration

1. `cmd` resolves the storage root via `internal/storage.ResolveRoot` (flag override → `MOCHI_STICKY_STORAGE` → legacy `mochi-sticky.yaml` → `.sticky/` default).
2. The resolved root is passed into board repositories (`internal/board`) and the wiki root helpers (`cmd/wiki_*`, `internal/mcp`, `internal/tui`).
3. `mochi-sticky hydrate` reports the hydrated paths so operators can validate storage/layout health.

### Task Model (Go Struct)

```go
type Task struct {
    ID           string      // Unique identifier (T-000001)
    UID          string      // UUID for collision-free IDs
    Title        string      // Task title
    Status       string      // Must match column key in config
    Priority     int         // 1 (high) to 3 (low)
    Tags         []string    // Classification tags
    CreatedAt    time.Time   // Creation timestamp
    Content      string      // Markdown description
    Dependencies []string    // Task IDs this depends on
    FilePath     string      // Physical file location
}
```

### Column Configuration (`.sticky/config.yaml`)

```yaml
board:
  columns:
    - name: "Backlog"
      key: "todo"
    - name: "In Progress"
      key: "doing"
    - name: "Review"
      key: "review"
    - name: "Finished"
      key: "done"
```

## Security & Constraints

- **Path Sandboxing:** Only files under `.sticky/` are eligible for operations.
- **Git Integrity:** `mochi-sticky` never performs `git commit` automatically.

## Key Design Patterns

### Repository Pattern
All file I/O goes through repository interfaces, enabling:
- Easy testing with mock implementations
- Consistent error handling
- Transaction-like semantics
- Caching opportunities

### Domain-Driven Design
Core business logic isolated in `internal/board/`:
- No dependencies on UI frameworks
- Pure functions where possible
- Clear domain models and boundaries

### The Elm Architecture (TUI)
Predictable state management:
- Immutable model updates
- Messages for all state changes
- Centralized update function
- Declarative view rendering

## File System Structure

```
.sticky/
├── mochi-sticky.yaml        # Tool configuration (paths, templates, editor)
├── boards/boards.yaml       # Board registry (active board marker)
├── boards/                  # All board data
│   └── <board-id>/
│       ├── config.yaml      # Board configuration and columns
│       ├── tasks/           # Active tasks
│       │   └── *.md         # Task files (YAML + Markdown)
│       └── archive/         # Archived tasks
│           └── tasks/*.md
├── wiki/                    # Documentation
│   ├── *.md                 # Wiki pages (YAML + Markdown)
│   └── templates/           # Wiki templates
└── adrs/                    # Architecture Decision Records
    └── *.md
```

## Concurrency & Safety

- `Repository` uses `sync.RWMutex` for concurrent access.
- TUI updates via messages only (no direct state mutation).
- MCP handles one request at a time per connection.

## Testing Strategy

- Table-driven tests for domain logic.
- Mock file systems for repository tests.
- Parser round-trip validation where possible.

## Extension Points

- Add commands in `cmd/` using domain layer functions.
- Expose new MCP tools in `internal/mcp/server.go`.
- Add export formats in `internal/wiki/export.go`.

## Performance Considerations

- Lazy loading of task content.
- Caching of parsed frontmatter.
- Batch operations where possible.
- Grep-based wiki search (indexing can be added later).

## Future Extensibility

- Graph view for wiki links.
- Webhooks on task transitions.
