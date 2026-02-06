---
title: Boards
slug: reference/boards
section: Reference
order: 22
tags:
    - boards
status: published
---
---
title: Boards
slug: reference/boards
section: Reference
order: 22
tags:
    - boards
status: published
---
# Boards

Boards organize tasks into separate workspaces with independent columns and configurations.

## Board Structure

Each board has its own directory:

```
.sticky/boards/
├── default/
│   ├── config.yaml
│   ├── tasks/
│   │   ├── T-000001.md
│   │   └── T-000002.md
│   └── archive/
│       └── tasks/
│           └── T-000099.md
├── project-alpha/
│   ├── config.yaml
│   └── tasks/
└── sprint-planning/
    ├── config.yaml
    └── tasks/
```

## Board Registry

Central registry at `.sticky/boards/boards.yaml`:

```yaml
active: default
boards:
  - id: default
    name: Default
    path: boards/default
    archived: false
    created: "2026-01-29"
  - id: project-alpha
    name: Project Alpha
    path: boards/project-alpha
    archived: false
    created: "2026-01-30"
  - id: sprint-planning
    name: Sprint Planning
    path: boards/sprint-planning
    archived: true
    created: "2026-01-30"
```

**Fields**:
- `active` — Currently selected board ID
- `boards` — List of all boards
  - `id` — Directory name (URL-safe)
  - `name` — Display name
  - `path` — Path under `.sticky/` to the board directory
  - `archived` — Whether board is hidden
  - `created` — Date the board was created

## Active Board

Only one board is active at a time.

**Commands run against the active board**:
```bash
mochi-sticky task list        # Lists tasks from active board
mochi-sticky task add "Title" # Creates task in active board
```

**Switch active board**:
```bash
mochi-sticky board use project-alpha
```

## Board Commands

### List Boards

```bash
# All boards
mochi-sticky board list

# Show active board
mochi-sticky board show
```

### Add Board

```bash
mochi-sticky board add "Project Beta"

# Generated board ID: "project-beta"
# Creates .sticky/boards/project-beta/
# Initializes config.yaml with default columns
```

### Use Board

```bash
# Switch to different board
mochi-sticky board use project-alpha

# Updates active_board in registry
```

### Rename Board

```bash
# Rename display name (ID stays same)
mochi-sticky board rename project-alpha "Alpha Project Reboot"

# Board ID: project-alpha (unchanged)
# Board Name: "Alpha Project Reboot" (updated)
```

### Archive Board

```bash
# Hide board from normal lists
mochi-sticky board archive project-alpha

# Board still exists, marked archived: true
```

### Delete Board

```bash
# Permanently remove board and all tasks
mochi-sticky board delete project-alpha --force

# ⚠️ DESTRUCTIVE: Cannot be undone
```

## Board Configuration

Each board has `.sticky/boards/<board-id>/config.yaml`:

```yaml
config_version: 1
next_id: 15
columns:
  - title: "Todo"
    key: "todo"
  - title: "In Progress"
    key: "doing"
  - title: "Done"
    key: "done"
context:
  scope: "Sprint 42"
  release: "v2.0.0"
```

See [Schema Reference](schema.md) for the full board config schema, and [Configuration](config.md) for detailed field guidance.

## Board Context

Track sprint/release metadata:

```yaml
context:
  scope: "Q1 2026 Features"
  release: "v3.0"
  target: "Roadmap-2026Q1"
  owners: ["team-alpha"]
  notes: "Security focus"
```

**Use cases**:
- Sprint planning
- Release tracking
- Team assignment
- Project notes

**Access via MCP**:
```json
{"method": "get_board_context"}
{"method": "update_board_context", "params": {
  "scope": "Sprint 43"
}}
```

## Multi-Board Workflows

### Project-Based Boards

```bash
# One board per project
mochi-sticky board add "Website Redesign"
mochi-sticky board add "Mobile App"
mochi-sticky board add "API Refactor"
```

### Sprint-Based Boards

```bash
# New board per sprint
mochi-sticky board add "Sprint 42"
mochi-sticky board add "Sprint 43"

# Archive old sprints
mochi-sticky board archive "sprint-41"
```

### Team-Based Boards

```bash
# Separate boards for teams
mochi-sticky board add "Backend Team"
mochi-sticky board add "Frontend Team"
mochi-sticky board add "DevOps Team"
```

## Board Best Practices

- **Use descriptive names**:
```bash
# Good
mochi-sticky board add "Q1 2026 Website Redesign"

# Less helpful
mochi-sticky board add "Board 2"
```

- **Archive instead of delete**:
```bash
# Preserve history
mochi-sticky board archive old-project

# Boards can be unarchived if needed
```

- **Document board purpose** in context:
```yaml
context:
  scope: "Customer Portal Rebuild"
  notes: "Replace legacy PHP with modern stack"
```

- **Customize columns per board**:
```yaml
# Development board
columns:
  - {name: "Backlog", key: "backlog"}
  - {name: "Doing", key: "doing"}
  - {name: "Review", key: "review"}
  - {name: "Done", key: "done"}

# Bug tracking board
columns:
  - {name: "Reported", key: "reported"}
  - {name: "Investigating", key: "investigating"}
  - {name: "Fixed", key: "fixed"}
  - {name: "Verified", key: "verified"}
```

- **Switch boards intentionally**:
```bash
# Check active board
mochi-sticky board show

# Switch with purpose
mochi-sticky board use project-alpha
```

## Board Isolation

**Boards are fully isolated**:
- Separate task numbering (each has own next_id)
- Independent columns/statuses
- Separate archive directories
- No task sharing between boards

**To reference tasks across boards**:
- Use task titles or UIDs in notes
- Cannot set cross-board dependencies
- Consider using tags for project tracking

## Troubleshooting

### Wrong Board Active

**Problem**: Commands affecting unexpected board

**Solution**:
```bash
# Check which board is active
mochi-sticky board show

# Switch to correct board
mochi-sticky board use project-alpha
```

### Board Not Found

**Error**: "board not found: xyz"

**Solution**:
```bash
# List available boards
mochi-sticky board list

# Check spelling and use exact ID
```

### Cannot Delete Board

**Error**: Deletion requires --force flag

**Reason**: Safety measure for destructive operation

**Solution**:
```bash
# Archive first (reversible)
mochi-sticky board archive old-board

# Or force delete (irreversible)
mochi-sticky board delete old-board --force
```

## Related

- [Tasks](tasks.md) — Task management
- [Configuration](config.md) — Board configuration details
- [CLI Usage](../user-guide/cli.md) — Command reference
- [TUI Usage](../user-guide/tui.md) — Visual board management
