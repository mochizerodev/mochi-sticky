---
title: Configuration
slug: reference/config
section: Reference
order: 20
tags:
    - config
status: published
---
# Configuration

mochi-sticky uses YAML configuration files to customize behavior, storage, and board structure.

## Storage Root Configuration

Control where mochi-sticky stores its data.

### .sticky/mochi-sticky.yaml

Place in your storage root (`.sticky/` by default):

```yaml
storage_root: .sticky
editor: "vim"
config_paths:
  boards: boards/boards.yaml
  adr: adrs/config.yaml
  wiki_index: wiki/_index.yaml
templates:
  root: .sticky/templates
  task: .sticky/templates/task.md
  board: .sticky/templates/board.yaml
  adr: .sticky/templates/adr.md
  wiki: .sticky/templates/wiki
```

### Config Paths

`config_paths` lets you relocate specific config files relative to `storage_root`:
- `boards` — board registry file (default: `boards/boards.yaml`)
- `adr` — ADR config file (default: `adrs/config.yaml`)
- `wiki_index` — wiki index file (default: `wiki/_index.yaml`)

### Configuration Precedence

1. **CLI Flag**: `--storage /path/to/storage`
2. **Environment Variable**: `MOCHI_STICKY_STORAGE=/path/to/storage`
3. **Legacy Config File**: `mochi-sticky.yaml` in the project root (storage root only)
4. **Default**: `.sticky/`

## Board Configuration

Each board has `.sticky/boards/<board-id>/config.yaml`:

```yaml
config_version: 1
next_id: 15
columns:
  - title: "Todo"
    key: "todo"
  - title: "Doing"
    key: "doing"
  - title: "Done"
    key: "done"
context:
  scope: "Sprint 42"
  release: "v2.0.0"
```

See [Schema Reference](schema.md) for the full board config schema and versioning details.

### Columns

**Required fields**:
- `title` — Display name (shown in TUI)
- `key` — Status identifier (used in task files)

**Task `status` field must match a column `key`**

### Common Patterns

**Simple workflow**:
```yaml
columns:
  - title: "Backlog"
    key: "backlog"
  - title: "Active"
    key: "active"
  - title: "Complete"
    key: "complete"
```

**Development workflow**:
```yaml
columns:
  - title: "To Do"
    key: "todo"
  - title: "Doing"
    key: "doing"
  - title: "Code Review"
    key: "review"
  - title: "Testing"
    key: "testing"
  - title: "Done"
    key: "done"
```

### Board Context

Optional metadata for sprint/release planning:

```yaml
context:
  scope: "Q1 2026 Features"
  release: "v3.0"
  target: "Roadmap-2026Q1"
  owners: ["team-alpha"]
  notes: "Security focus"
```

**Access via MCP**: `get_board_context`, `update_board_context`

## Editor Configuration

In `.sticky/mochi-sticky.yaml`:

```yaml
editor: "vim"
```

**Precedence**:
1. CLI flag: `--editor "vim"`
2. Environment: `$MOCHI_EDITOR`
3. Environment: `$EDITOR`
4. Config file (`editor` in `.sticky/mochi-sticky.yaml`)
5. Default: `nano`

## Board Registry

`.sticky/boards/boards.yaml`:

```yaml
active: "default"
boards:
  - id: "default"
    name: "Default"
    path: "boards/default"
    archived: false
    created: "2026-01-29"
```

## ADR Configuration

`.sticky/adrs/config.yaml`:

```yaml
config_version: 1
next_id: 5
columns:
  - key: proposed
    title: Proposed
  - key: accepted
    title: Accepted
  - key: deprecated
    title: Deprecated
  - key: superseded
    title: Superseded
```

## Validation

```bash
mochi-sticky hydrate
mochi-sticky hydrate --json --pretty
```

## Best Practices

- **Version control** `.sticky/mochi-sticky.yaml`
- **Use descriptive column names**
- **Keep keys simple** (lowercase, alphanumeric)
- **Backup before major changes**: `cp -r .sticky .sticky.backup`

## Related

- [Tasks](tasks.md) — Task file structure
- [Boards](boards.md) — Board organization
- [CLI Usage](../user-guide/cli.md) — Command reference
