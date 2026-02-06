---
title: CLI Usage
slug: user-guide/cli
section: User Guide
order: 10
tags:
    - cli
status: published
---
---
title: CLI Usage
slug: user-guide/cli
section: User Guide
order: 10
tags:
    - cli
status: published
---
# CLI Usage

Complete command reference for mochi-sticky CLI.

## Global Flags

```bash
--storage <path>   # Override storage root
--help             # Show help
--version          # Show version
```

## Task Management

### List Tasks

```bash
mochi-sticky task list
mochi-sticky task list --status todo
mochi-sticky task list --tag backend
mochi-sticky task list --sort priority
mochi-sticky task list --from 2026-01-01 --to 2026-01-31
```

**Flags**:
- `--status <key>` — Filter by status
- `--tag <name>` — Filter by tag
- `--tag-mode <any|all>` — Tag match mode
- `--sort <field>` — Sort by: priority, created, title
- `--desc` — Reverse sort order
- `--from <date>` — Filter by creation date
- `--to <date>` — End date for range

### Add Task

```bash
mochi-sticky task add "Task title"
mochi-sticky task add "Fix bug" --priority 1 --tags backend,bugfix
```

**Flags**:
- `--priority <1|2|3>` — Set priority (default: 2)
- `--tags <tag1,tag2>` — Comma-separated tags
- `--status <key>` — Initial status (default: todo)

### Show Task

```bash
mochi-sticky task show T-000042
mochi-sticky task show T-000042 --metadata
```

### Move Task

```bash
mochi-sticky task move T-000042 doing
mochi-sticky task move T-000042 done
```

### Update Priority

```bash
mochi-sticky task priority T-000042 1
```

### Manage Dependencies

```bash
# Set dependencies
mochi-sticky task deps T-000042 --set T-000040,T-000041

# View dependencies
mochi-sticky task deps T-000042

# List ready tasks (no blockers)
mochi-sticky task ready
```

### Archive Tasks

```bash
# Archive single task
mochi-sticky task archive task T-000042

# Archive by date
mochi-sticky task archive before 2025-12-31

# List archived
mochi-sticky task archive list

# Restore from archive
mochi-sticky task archive restore T-000042

# Delete archived (permanent)
mochi-sticky task archive delete T-000042 --force
```

### Delete Task

```bash
mochi-sticky task delete T-000042 --force
```

## Board Management

### List Boards

```bash
mochi-sticky board list
mochi-sticky board list --include-archived
```

### Show Active Board

```bash
mochi-sticky board show
```

### Add Board

```bash
mochi-sticky board add "Project Alpha"
```

### Switch Board

```bash
mochi-sticky board use project-alpha
```

### Rename Board

```bash
mochi-sticky board rename project-alpha "Alpha Reboot"
```

### Archive Board

```bash
mochi-sticky board archive project-alpha
```

### Delete Board

```bash
mochi-sticky board delete old-board --force
```

## Wiki Management

### List Pages

```bash
mochi-sticky wiki list
mochi-sticky wiki list --section "User Guide"
mochi-sticky wiki list --tag tutorial
```

### View Page

```bash
mochi-sticky wiki view getting-started/install
```

### Create Page

```bash
mochi-sticky wiki create \
  --title "New Guide" \
  --slug user-guide/new-guide \
  --section "User Guide" \
  --order 40
```

### Edit Page

```bash
mochi-sticky wiki edit user-guide/cli
```

**Uses editor from**:
1. `--editor` flag
2. `$MOCHI_EDITOR` environment
3. `$EDITOR` environment
4. `.sticky/mochi-sticky.yaml` (`editor`)
5. Default: `nano`

### Search Wiki

```bash
mochi-sticky wiki search "authentication"
mochi-sticky wiki search "API" --section Reference
```

### Generate Index

```bash
mochi-sticky wiki index generate
```

### Export Wiki

```bash
# Single Markdown file
mochi-sticky wiki export markdown --output docs.md

# PDF (requires Pandoc)
mochi-sticky wiki export pdf --output docs.pdf

# Multi-root merge
mochi-sticky wiki export multi-root \
  --roots /path1/.sticky/wiki,/path2/.sticky/wiki \
  --output merged.md
```

See [Markdown Export](../export/markdown.md) and [PDF Export](../export/pdf.md) for details.

## ADR Management

### List ADRs

```bash
mochi-sticky adr list
mochi-sticky adr list --status accepted
mochi-sticky adr list --tag security
```

### Create ADR

```bash
mochi-sticky adr create "Use PostgreSQL for primary database"
mochi-sticky adr create "Migrate to GraphQL" --tags api,architecture
mochi-sticky adr create "Replace X with Y" --supersedes 4
```

### View ADR

```bash
mochi-sticky adr view 4
mochi-sticky adr view 4 --metadata
```

### Edit ADR

```bash
mochi-sticky adr edit 4
mochi-sticky adr edit 4 --editor vim
```

### Move ADR

```bash
mochi-sticky adr move 4 10
```

### List Statuses

```bash
mochi-sticky adr statuses
```

### Lint ADRs

```bash
mochi-sticky adr lint
```

## Storage & Initialization

### Initialize Storage

```bash
mochi-sticky init
mochi-sticky init --storage /custom/path
mochi-sticky init --force  # Reinitialize
```

### Validate Storage

```bash
mochi-sticky hydrate
mochi-sticky hydrate --json --pretty
```

## TUI

### Launch TUI

```bash
mochi-sticky tui
mochi-sticky tui --board project-alpha
```

### Set Editor

```bash
mochi-sticky tui --set-editor "code --wait"
```

## MCP Server

### Start MCP Server

```bash
mochi-sticky mcp
```

See [MCP Usage](mcp.md) for integration details.

## Examples

### Daily Workflow

```bash
# Morning: check tasks
mochi-sticky task list --status doing

# Start working on task
mochi-sticky task move T-000042 doing

# Add new urgent task
mochi-sticky task add "Fix prod issue" --priority 1

# End of day: archive completed
mochi-sticky task list --status done
mochi-sticky task archive task T-000040
```

### Project Setup

```bash
# Create project board
mochi-sticky board add "Website Redesign"
mochi-sticky board use website-redesign

# Add initial tasks
mochi-sticky task add "Design mockups" --tags design
mochi-sticky task add "Setup repo" --tags dev
mochi-sticky task add "Plan architecture" --tags architecture

# View board
mochi-sticky tui
```

### Documentation Updates

```bash
# Create new wiki page
mochi-sticky wiki create \
  --title "API Authentication" \
  --slug user-guide/auth \
  --section "User Guide" \
  --order 50

# Edit page
mochi-sticky wiki edit user-guide/auth

# Export to PDF
mochi-sticky wiki export pdf --output user-guide.pdf
```

## Tips

- **Use tab completion** (if shell configured)

- **Chain commands** with `&&`:
```bash
mochi-sticky task add "New feature" && mochi-sticky tui
```

- **Check active board** before task operations:
```bash
mochi-sticky board show
```

- **Use `--help` on any command**:
```bash
mochi-sticky task --help
mochi-sticky task add --help
```

- **JSON output for automation**:
```bash
mochi-sticky task list --json
mochi-sticky hydrate --json --pretty
```

## Related

- [TUI Usage](tui.md) — Interactive interface
- [MCP Usage](mcp.md) — AI agent integration
- [Tasks Reference](../reference/tasks.md) — Task structure
- [Configuration](../reference/config.md) — Configuration files
