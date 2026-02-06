---
title: Tasks
slug: reference/tasks
section: Reference
order: 21
tags:
    - tasks
status: published
---
---
title: Tasks
slug: reference/tasks
section: Reference
order: 21
tags:
    - tasks
status: published
---
# Tasks

Tasks are the core units of work in mochi-sticky. Each task is a Markdown file with YAML frontmatter.

## File Location

**Active**: `.sticky/boards/<board-id>/tasks/<task-id>.md`  
**Archived**: `.sticky/boards/<board-id>/archive/tasks/<task-id>.md`

## Task Structure

```markdown
---
id: "T-000042"
uid: "c2f0b4d1a3a14a1b91ad2e9f0e2c3f5d"
title: "Implement user authentication"
status: "doing"
priority: 1
tags:
  - backend
  - security
created: 2026-02-04
dependencies:
  - T-000040
---
# Task Description

Implementation details here.

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2
```

## Frontmatter Fields

### id (required)

Unique identifier: `T-NNNNNN` (6-digit zero-padded)

**Examples**: `T-000001`, `T-000042`

### title (required)

Short description (keep under 80 chars).

**Best practices**: Start with action verb, be specific.

### status (required)

Current state — **must match board column `key`**:

```yaml
# Board config columns
columns:
  - key: "todo"   # ← Valid status
  - key: "doing"  # ← Valid status
  - key: "done"   # ← Valid status
```

### priority (optional)

Urgency level:
- `1` — High priority (urgent)
- `2` — Normal (default)
- `3` — Low priority

### tags (optional)

Classification labels:

```yaml
tags: [backend, api]
tags:
  - frontend
  - ui
```

**Common tags**:
- Component: `backend`, `frontend`, `database`
- Type: `feature`, `bugfix`, `refactor`
- Area: `auth`, `payments`, `search`

### created (required)

Creation date: `YYYY-MM-DD`

Auto-generated when creating tasks.

### dependencies (optional)

Blocking tasks:

```yaml
dependencies:
  - T-000040
  - T-000041
```

**Usage**:
```bash
mochi-sticky task deps T-000042 --set T-000040,T-000041
mochi-sticky task ready  # List tasks with no blockers
```

## Task Lifecycle

### Creation

```bash
mochi-sticky task add "Title" --tags backend --priority 1
```

### Status Changes

```bash
mochi-sticky task move T-000042 doing
mochi-sticky task move T-000042 done
```

### Archiving

```bash
mochi-sticky task archive task T-000042
mochi-sticky task archive before 2025-12-31
```

### Restoration

```bash
mochi-sticky task archive restore T-000042
```

## Filtering

**By status**: `mochi-sticky task list --status todo`  
**By priority**: `mochi-sticky task list --sort priority`  
**By tags**: `mochi-sticky task list --tag backend --tag-mode any`  
**By date**: `mochi-sticky task list --from 2026-01-01`

## MCP Integration

```json
{"method": "create_task", "params": {
  "title": "New feature",
  "priority": 1,
  "tags": ["backend"]
}}

{"method": "update_task_status", "params": {
  "id": "T-000042",
  "status": "doing"
}}
```

See [MCP Usage](../user-guide/mcp.md) for complete API.

## Best Practices

- **Use descriptive titles**: "Add email validation" not "Fix bug"  
- **Tag consistently** across your team  
- **Reserve priority 1** for truly urgent work  
- **Document acceptance criteria** in task body  
- **Archive completed work** periodically

## Related

- [Boards](boards.md) — Board organization
- [Configuration](config.md) — Status columns
- [CLI Usage](../user-guide/cli.md) — Task commands
- [TUI Usage](../user-guide/tui.md) — Interactive management
