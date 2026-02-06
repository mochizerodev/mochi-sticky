---
title: mochi-sticky Wiki
slug: home
section: Home
order: 0
tags:
    - overview
status: published
---
# mochi-sticky Wiki

Welcome to the **mochi-sticky** documentation — the sticky-note project manager for developers.

## What is mochi-sticky?

mochi-sticky is a file-based project management tool that lives in your Git repository. It treats Markdown files as "sticky notes" on a Kanban board, giving you:

- **CLI** for rapid task management
- **TUI** for an interactive Kanban experience
- **MCP** for AI agent integration
- **Wiki** for comprehensive documentation
- **ADR** for tracking architectural decisions

All data lives in `.sticky/` as Markdown files with YAML frontmatter — no external database required.

## Navigation

### Getting Started
- [Installation](getting-started/install.md) — Build from source and initialize
- [Quickstart](getting-started/quickstart.md) — Create your first task in 30 seconds

### User Guide
- [CLI Usage](user-guide/cli.md) — Command reference for daily workflow
- [TUI Usage](user-guide/tui.md) — Keyboard shortcuts and interactive board navigation
- [Wiki Basics](user-guide/wiki.md) — Document management and organization
- [MCP Usage](user-guide/mcp.md) — Integrate with AI agents via Model Context Protocol
- [Use Cases](user-guide/use-cases.md) — Workflow examples and patterns

### Reference
- [Configuration](reference/config.md) — Storage root, templates, and settings
- [Tasks](reference/tasks.md) — Task schema and metadata
- [Boards](reference/boards.md) — Multi-board management and context
- [Wiki Schema](reference/wiki-schema.md) — Wiki page structure and sections
- [Templates](reference/templates.md) — Customizing task, ADR, and wiki templates

### Export & Integration
- [Markdown Export](export/markdown.md) — Export wiki to single markdown file
- [PDF Export](export/pdf.md) — Generate PDF documentation
- [Multi-root Export](export/multi-root.md) — Combine multiple wiki sources

### Development
- [Architecture](development/architecture.md) — System design and components
- [Contributing](development/contributing.md) — Guidelines for contributors
- [Testing](development/testing.md) — Test strategy and running tests

### Troubleshooting
- [Common Issues](troubleshooting/common-issues.md) — Solutions to frequent problems

## Key Features

### Multi-Board Support
Organize tasks across multiple boards with an active board registry. Each board has its own configuration, columns, and task collection.

### Flexible Task Metadata
- **Status** — Configurable columns (todo → doing → done)
- **Priority** — 1 (highest) to 3 (lowest)
- **Tags** — Organize with multiple labels
- **Dependencies** — Track task relationships
- **Created Date** — Automatic timestamp

### Architecture Decision Records (ADRs)
Document important architectural decisions with status tracking (proposed → accepted → deprecated).

### Agent Integration
The MCP server exposes JSON-RPC tools so AI agents can:
- List, create, update, and delete tasks
- Manage boards and configurations
- Read and write wiki pages
- Search across documentation

## Quick Commands

```bash
# Initialize mochi-sticky
./mochi-sticky init

# Add a task
./mochi-sticky task add "Implement login" --tags backend --priority 1

# Launch the TUI
./mochi-sticky tui

# Start MCP server
./mochi-sticky mcp

# Create a wiki page
./mochi-sticky wiki create "Getting Started Guide" --section guides
```

## Philosophy

mochi-sticky follows these principles:

1. **Data is Code** — Everything lives in your Git repo
2. **Markdown First** — Human-readable, version-controlled files
3. **No External Dependencies** — Single binary, no database
4. **Developer-Friendly** — CLI, TUI, and API-first design
5. **Agent-Ready** — Built for AI collaboration from the ground up

## Next Steps

New to mochi-sticky? Start here:
1. [Install mochi-sticky](getting-started/install.md)
2. Run through the [Quickstart](getting-started/quickstart.md)
3. Explore the [CLI Usage](user-guide/cli.md) guide
4. Launch the [TUI](user-guide/tui.md) for interactive management

For AI agent integration, check out [MCP Usage](user-guide/mcp.md).
