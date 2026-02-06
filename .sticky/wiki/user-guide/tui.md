---
title: TUI Usage
slug: user-guide/tui
section: User Guide
order: 11
tags:
    - tui
status: published
---
---
title: TUI Usage
slug: user-guide/tui
section: User Guide
order: 11
tags:
    - tui
status: published
---
# TUI Usage

Interactive terminal interface for visual task management.

## Launch TUI

```bash
mochi-sticky tui
mochi-sticky tui --board project-alpha
```

## Interface Overview

```
┌─ To Do ─────────┬─ In Progress ───┬─ Done ──────────┐
│ T-000001        │ T-000003        │ T-000005        │
│ Fix login bug   │ Add API auth    │ Update docs     │
│ Priority: 1     │ Priority: 2     │ Priority: 2     │
│                 │                 │                 │
│ T-000002        │ T-000004        │                 │
│ Update README   │ Design mockups  │                 │
│ Priority: 2     │ Priority: 1     │                 │
└─────────────────┴─────────────────┴─────────────────┘

[a]dd  [x]actions  [z]archive  [q]uit
```

## Navigation

### Between Columns

- `h` or `←` — Move left
- `l` or `→` — Move right

### Between Tasks

- `j` or `↓` — Move down
- `k` or `↑` — Move up
- `g` — Jump to top
- `G` — Jump to bottom

### Task Selection

- `Enter` — View task details
- `Esc` — Return to board view

## Task Operations

### Add Task

Press `a`:

```
┌─ Add Task ──────────────────────────┐
│ Title: Fix authentication bug       │
│ Priority: [1-High 2-Normal 3-Low]: 1│
│ Tags (comma-separated): auth,bugfix │
│ Status: todo                        │
└─────────────────────────────────────┘
```

**Fields**:
- Title (required)
- Priority (1, 2, or 3)
- Tags (comma-separated)
- Status (must match column key)

### Task Actions Menu

Press `x` on selected task:

```
┌─ Task Actions: T-000042 ────────────┐
│ [m] Move to different status        │
│ [p] Change priority                 │
│ [t] Edit tags                       │
│ [d] Add/edit dependencies           │
│ [e] Edit in external editor         │
│ [a] Archive task                    │
│ [x] Delete task                     │
│ [q] Cancel                          │
└─────────────────────────────────────┘
```

### Move Task

From actions menu, press `m`:

```
┌─ Move Task ─────────────────────────┐
│ Select new status:                  │
│   [1] todo                          │
│   [2] doing                         │
│ → [3] review                        │
│   [4] done                          │
└─────────────────────────────────────┘
```

Use `↑`/`↓` or number keys to select.

### Change Priority

From actions menu, press `p`:

```
Select priority:
  [1] High (urgent)
  [2] Normal
  [3] Low
```

### Edit Task

From actions menu, press `e`:

Opens task in configured editor:
1. `$MOCHI_EDITOR`
2. `$EDITOR`
3. `.sticky/mochi-sticky.yaml` (`editor`)
4. Default: `nano`

**Set editor**:
```bash
export MOCHI_EDITOR="code --wait"
# or
mochi-sticky tui --set-editor "vim"
```

## Archive Browser

Press `z` to view archived tasks:

```
┌─ Archive Browser ───────────────────┐
│ T-000010 | Completed feature        │
│ T-000015 | Fixed critical bug       │
│ T-000020 | Updated documentation    │
│                                     │
│ [r]estore  [d]elete  [q]uit        │
└─────────────────────────────────────┘
```

**Actions**:
- `r` — Restore to active board
- `d` — Permanently delete (requires confirmation)
- `q` — Return to board

## Board Switching

Press `b` to change active board:

```
┌─ Select Board ──────────────────────┐
│ → default (Default Board)           │
│   project-alpha (Project Alpha)     │
│   website (Website Redesign)        │
│                                     │
│ [Enter] to select  [q] to cancel   │
└─────────────────────────────────────┘
```

## Filtering & Search

Press `/` to search tasks:

```
┌─ Search ────────────────────────────┐
│ Query: authentication               │
│                                     │
│ Matches:                            │
│ T-000042 | Implement auth           │
│ T-000055 | Fix auth redirect        │
└─────────────────────────────────────┘
```

**Search filters**:
- Title text
- Task ID
- Tags

Press `Esc` to clear search.

## Keyboard Shortcuts Reference

### Navigation
- `h`/`←` — Previous column
- `l`/`→` — Next column
- `j`/`↓` — Next task
- `k`/`↑` — Previous task
- `g` — Top of list
- `G` — Bottom of list

### Actions
- `a` — Add new task
- `x` — Task actions menu
- `z` — Archive browser
- `b` — Switch board
- `/` — Search tasks
- `Enter` — View task details
- `Esc` — Cancel/close

### Global
- `?` — Show help
- `r` — Refresh view
- `q` — Quit TUI

## Task Details View

Press `Enter` on a task:

```
┌─ Task Details: T-000042 ────────────┐
│ Title: Implement user authentication│
│ Status: doing                       │
│ Priority: 1 (High)                  │
│ Tags: backend, security, auth       │
│ Created: 2026-02-04                 │
│ Dependencies: T-000040, T-000041    │
│                                     │
│ ─── Description ───                 │
│ Implement JWT-based authentication  │
│ for the API.                        │
│                                     │
│ ## Acceptance Criteria              │
│ - [ ] Token generation works        │
│ - [ ] Session management            │
│                                     │
│ [e]dit  [x]actions  [Esc]back      │
└─────────────────────────────────────┘
```

## Color Coding

Tasks are color-coded by priority:

- 🔴 **Red** — Priority 1 (High)
- 🟡 **Yellow** — Priority 2 (Normal)
- 🟢 **Green** — Priority 3 (Low)

## Tips & Best Practices

- **Use keyboard shortcuts**: Much faster than menus

- **Regular archiving**: Press `z` to review and clean up

- **Visual overview**: TUI is best for seeing full board state

- **Quick status changes**: `x` → `m` → select column

- **Batch operations**: Use CLI for bulk actions, TUI for individual tasks

## Common Workflows

### Morning Standup

1. Launch TUI: `mochi-sticky tui`
2. Review "In Progress" column
3. Move completed tasks to "Done"
4. Select next task and move to "Doing"
5. Check high-priority tasks (red highlights)

### Sprint Planning

1. Switch to sprint board: Press `b`
2. Add tasks: Press `a` repeatedly
3. Set priorities appropriately
4. Organize by dependencies
5. Archive old completed tasks: Press `z`

### Bug Triage

1. Add urgent bug: Press `a`
2. Set priority to 1 (High)
3. Tag as `bugfix`, `critical`
4. Move to "Doing" immediately
5. Update via actions menu as you work

## Troubleshooting

### TUI Doesn't Display Correctly

**Problem**: Garbled output or layout issues

**Solutions**:
1. Increase terminal size (minimum 80x24)
2. Use a terminal with full Unicode support
3. Check `$TERM` environment variable

### Can't Edit Tasks

**Problem**: Editor doesn't open

**Solutions**:
1. Set editor explicitly:
   ```bash
   mochi-sticky tui --set-editor "nano"
   ```
2. Check editor installation:
   ```bash
   which nano vim code
   ```

### Tasks Don't Refresh

**Solution**: Press `r` to manually refresh

## Related

- [CLI Usage](cli.md) — Command-line interface
- [Tasks Reference](../reference/tasks.md) — Task structure
- [Configuration](../reference/config.md) — Board columns
