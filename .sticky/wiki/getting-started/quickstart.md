---
title: Quickstart
slug: getting-started/quickstart
section: Getting Started
order: 2
tags:
    - quickstart
status: published
---
# Quickstart

Get started with mochi-sticky in 5 minutes.

## 1. Install

```bash
go build -o mochi-sticky
sudo cp mochi-sticky /usr/local/bin/
```

## 2. Initialize

```bash
cd your-project
mochi-sticky init
```

Creates .sticky/ directory with default board.

## 3. Create Your First Task

```bash
mochi-sticky task add "My first task" --tags onboarding --priority 1
```

## 4. List Tasks

```bash
mochi-sticky task list --sort priority
```

## 5. Move Task

```bash
mochi-sticky task move T-000001 doing
mochi-sticky task move T-000001 done
```

## 6. Launch TUI

```bash
mochi-sticky tui
```

Keyboard shortcuts:
- h/l: switch columns
- j/k: navigate tasks
- a: add task
- m: move task
- enter: task details
- q: quit

## 7. Create Documentation

```bash
mochi-sticky wiki create "Project Guide" --section guides
mochi-sticky wiki list
mochi-sticky wiki view home
```

## 8. Multiple Boards

```bash
mochi-sticky board add "Sprint 1"
mochi-sticky board use board-2
mochi-sticky task add "Sprint task"
mochi-sticky board list
```

## Next Steps

- [CLI Usage](../user-guide/cli.md) — Full command reference
- [TUI Usage](../user-guide/tui.md) — Interactive interface guide
- [Configuration](../reference/config.md) — Customize settings
- [MCP Usage](../user-guide/mcp.md) — AI agent integration

## Pro Tips

```bash
# Filter tasks by tags
mochi-sticky task list --tag backend --tag-mode all

# Archive completed work
mochi-sticky task archive before 2026-01-31

# Export wiki to PDF
mochi-sticky wiki export --format pdf --output docs.pdf

# Validate configuration
mochi-sticky hydrate
```

Welcome to mochi-sticky! 🍡
