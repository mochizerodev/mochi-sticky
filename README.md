# mochi-sticky

<p align="center">
	<img src="assets/large-sticky.svg" alt="mochi-sticky" width="128" height="128">
</p>

> **The Sticky-Note Project Manager for Developers.**
> 
> A file-based project management tool that lives in your Git repo. Manage tasks, boards, wiki documentation, and Architecture Decision Records — all in Markdown.

[![Go Version](https://img.shields.io/badge/Go-1.23.12+-00ADD8?style=flat&logo=go)](https://go.dev)
[![CI](https://github.com/mochizero0/mochi-sticky/actions/workflows/ci.yml/badge.svg)](https://github.com/mochizero0/mochi-sticky/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mochizero0/mochi-sticky)](https://goreportcard.com/report/github.com/mochizero0/mochi-sticky)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> **Status: BETA (pre-1.0)**  
> The CLI/TUI and file formats may change without notice. Please expect breaking changes until a `v1.0.0` release.

## Support

If `mochi-sticky` saves you time and you’d like to support ongoing development, you can sponsor me on GitHub: [github.com/sponsors/mochizerodev](https://github.com/sponsors/mochizerodev)

## Platform Support Matrix

The project is currently only tested on Linux (CI). Other platforms may work, but are not guaranteed.

| mochi-sticky version | Linux | macOS | Windows |
| --- | --- | --- | --- |
| `main` / `dev` | Tested | Not tested | Not tested |

## AI Assistance

This project may use AI-assisted tooling during development. All changes are reviewed by maintainers and must pass CI.

## ✨ Features

- 📋 **Multi-Board Kanban** — Organize tasks across projects
- 🖥️ **CLI & TUI** — Command-line power with interactive UI
- 🤖 **MCP Integration** — AI agent support via Model Context Protocol
- 📚 **Built-in Wiki** — Document alongside your code
- 🏗️ **ADR Support** — Track architectural decisions
- 📦 **No Database** — Everything is Markdown + YAML
- 🔄 **Git-Friendly** — Version control your workflow

<!-- Favicons: they live in assets/brand/favicons/; generate with `scripts/generate-favicons.sh` -->

## 1. Project Vision

`mochi-sticky` keeps task management and documentation alongside your code. Each task is a Markdown file, organized into boards and columns, and managed through a fast CLI or an interactive TUI.

**Perfect for:**
- Solo developers tracking personal projects
- Small teams managing sprints
- Open source projects coordinating work
- Anyone who wants task management in Git

## 2. Quick Start

```bash
# Build
go build -o mochi-sticky

# Install (optional)
sudo cp mochi-sticky /usr/local/bin/

# Initialize in your project
cd your-project
mochi-sticky init

# Create your first task
mochi-sticky task add "Implement feature X" --tags backend --priority 1

# Launch the TUI
mochi-sticky tui

# Or use the CLI
mochi-sticky task list --sort priority
mochi-sticky task move T-000001 doing
```

📖 **Built-in Wiki** — Stored under `.sticky/wiki` and viewable with `mochi-sticky wiki view home`

🤖 **AI Agent Support** — Use with GitHub Copilot via MCP. See [AGENTS.md](AGENTS.md) for setup.

### Documentation (Internal Wiki)

- Wiki Home: [`.sticky/wiki/home.md`](.sticky/wiki/home.md)

## Development Dependencies

- `golangci-lint` (optional, for linting)
- `pre-commit` (recommended, runs lint before commits)

## 3. Capabilities

- Multi-board support with an active board registry.
- Task metadata: status, priority (1-3), tags, created date, plus Markdown content.
- ADRs (Architecture Decision Records) stored as Markdown + YAML frontmatter.
- CLI for task/board CRUD, filtering, sorting, archiving, and delete confirmations.
- TUI with a board selector, Kanban view, task detail editor, status picker, archive browser, plus an ADR Kanban view.
- External editor integration via `$EDITOR` or a configured override.
- **MCP Server**: 37 tools for AI agents (GitHub Copilot, Claude, etc.) to manage tasks, boards, and wiki. See [AGENTS.md](AGENTS.md).

## 3. Project Structure

```text
mochi-sticky/
├── .sticky/                # Default data directory (configurable)
│   ├── mochi-sticky.yaml   # Config (paths/templates)
│   ├── boards/boards.yaml  # Board registry
│   ├── adrs/               # Architecture Decision Records (ADRs)
│   └── boards/             # Board data (config, tasks, archives)
│       └── <board-id>/
│           ├── config.yaml
│           ├── tasks/
│           └── archive/
│               └── tasks/
├── cmd/                    # Cobra commands
├── internal/
│   ├── core/               # Parsing, storage, and business logic
│   └── tui/                # Bubble Tea models and views
├── go.mod
└── README.md
```

## 4. Data Schema (Task File)

Every task is a file in `.sticky/boards/<board-id>/tasks/*.md`. Status values are config-driven and must match a column `key` in `.sticky/boards/<board-id>/config.yaml`.

```markdown
---
id: "T-000001"
uid: "c2f0b4d1a3a14a1b91ad2e9f0e2c3f5d"
title: "Implement Auth Logic"
status: "todo"
priority: 1
tags: [backend, security]
created: 2026-01-29
---

# Task Description
Describe the task here. You can use standard Markdown.
```

Priority rules: `1` is highest, `3` is lowest. Unset priority defaults to `2`.

## 5. Storage Root Configuration

By default, mochi-sticky stores data under `.sticky/`. You can override the storage root in three ways (highest to lowest priority):

1. CLI flag: `--storage /path/to/store`
2. Environment variable: `MOCHI_STICKY_STORAGE=/path/to/store`
3. Legacy YAML config (repo root): `mochi-sticky.yaml`

Default config file location: `.sticky/mochi-sticky.yaml`

Example config file:
```yaml
storage_root: .sticky
editor: "vim"
config_paths:
  boards: boards/boards.yaml
  adr: adrs/config.yaml
  wiki_index: wiki/_index.yaml
templates:
  root: .sticky/templates
  adr: .sticky/templates/adr
  task: .sticky/templates/task
  board: .sticky/templates/board
  wiki: .sticky/templates/wiki
  wiki_pdf: .sticky/templates/wiki/wiki_pdf_template.tex
```

If no override is set, `.sticky/` is used. Invalid or non-existent paths fail fast unless you are running `mochi-sticky init`, which can create the directory.

`mochi-sticky init` also seeds default templates (ADR, task, board, wiki) into the configured template directories when they are empty. Defaults are embedded in the binary, so no extra repo template directory is required at runtime. If you have legacy templates under `.sticky/wiki/templates` or `.sticky/adrs/templates`, running `mochi-sticky init` will copy them into the unified template locations when those directories are empty.

`templates.wiki_pdf` is the preferred PDF template setting; the legacy `pdf_template` key is still supported.

## 6. CLI Usage

Root:
- `mochi-sticky` (no args): show CLI help
- `mochi-sticky init`: scaffold `.sticky` and default board
- `mochi-sticky hydrate`: validate storage/config and print a summary (use `--json [--pretty]` for automation)
- `mochi-sticky tui`: launch the TUI

Tasks:
- `mochi-sticky task add "Title" [--tags tag1,tag2] [--priority 1|2|3]`
- `mochi-sticky task list [--status todo] [--title "foo"] [--tag tag] [--tag-mode any|all] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--sort status|created|title|priority] [--desc]`
- `mochi-sticky task show <id>`
- `mochi-sticky task move <id> <status>`
- `mochi-sticky task deps <id> [--set T-000123,T-000456]` (view/set dependencies)
- `mochi-sticky task ready` (list tasks whose dependencies are satisfied)
- `mochi-sticky task statuses`
- `mochi-sticky task priority <id> <priority>`
- `mochi-sticky task delete <id> [--force]`
- `mochi-sticky task archive task <id> [--force]`
- `mochi-sticky task archive before <YYYY-MM-DD> [--force]`
- `mochi-sticky task archive list`
- `mochi-sticky task archive restore <id> [--force]`
- `mochi-sticky task archive delete <id> [--force]`

Task detail outputs (both `mochi-sticky task show` and the TUI detail view) now surface the board title near the task header.

Boards:
- `mochi-sticky board list`
- `mochi-sticky board show <id>`
- `mochi-sticky board add "Name"`
- `mochi-sticky board rename <id> "New Name"`
- `mochi-sticky board use <id>`
- `mochi-sticky board archive <id> [--force]`
- `mochi-sticky board delete <id> [--force]`
- `mochi-sticky board show <id>` now prints the context block (scope, release target, owners, notes).

Board context metadata (scope, release target, owners, notes) is stored in `.sticky/boards/<id>/config.yaml`. Use the MCP calls `update_board_context` / `get_board_context` to keep it in sync with CLI/TUI views.

Wiki:
- `mochi-sticky wiki create "Title" [--slug slug] [--section Section] [--order N] [--tags tag1,tag2] [--status draft|published|archived] [--template name]`
- `mochi-sticky wiki list`
- `mochi-sticky wiki view <slug>`
- `mochi-sticky wiki edit <slug> [--editor "cmd"]`
- `mochi-sticky wiki search <query>`
- `mochi-sticky wiki list --include-templates`
- `mochi-sticky wiki search <query> --include-templates`
- `mochi-sticky wiki manifest`
- `mochi-sticky wiki export --format md [--output path]`
- `mochi-sticky wiki export --format pdf [--output path] [--title "Title"] [--author "Name"] [--template path]`
- `mochi-sticky wiki export --format md --root /path/to/wiki[:prefix] --prefix main`
- `mochi-sticky wiki delete <slug> [--update-index]`
- `mochi-sticky wiki index [--include-templates] [--write=false] [--output path]`
- `mochi-sticky wiki list --status <status>`
- `mochi-sticky wiki search <query> --status <status>`
- `mochi-sticky wiki lint`

Destructive actions prompt by default; pass `--force` to skip confirmation.

ADRs:
- `mochi-sticky adr create "Title" [--status proposed] [--date YYYY-MM-DD] [--tags tag1,tag2] [--links item1,item2] [--body -]`
- `mochi-sticky adr list [--status accepted] [--tags foo,bar] [--query keyword] [--since YYYY-MM-DD] [--until YYYY-MM-DD]`
- `mochi-sticky adr view <id>`
- `mochi-sticky adr edit <id> [--editor "code --wait"]`
- `mochi-sticky adr move <id> <status>`
- `mochi-sticky adr statuses`
- `mochi-sticky adr lint`

## 7. TUI Usage

Run `mochi-sticky tui` to enter the interactive board.

Board view:
- `h/l`: switch columns
- `j/k`: move between tasks
- `a`: add a new task (tab between fields; `1-3` sets priority)
- `m`: move task forward
- `M`: move task back
- `x`: task actions menu
- `z`: archive browser
- `b`: boards selector
- `i`: board detail (shows description)
- `ctrl+r` or `F5`: refresh board & tasks (keeps selection when possible)
- `q`: quit
- `enter`: open task detail
- Context block (scope/release/owners/notes) sits above the columns so board-wide metadata stays visible.

Task detail:
- `tab`: cycle fields (Title, Status, Priority, Tags, Description)
- `enter`: edit the selected field
- `e`: open in external editor
- `a`: archive task
- `d`: delete task
- `x`: actions menu
- `esc`: back
- Board title stays visible at the top of the detail output so you always know which board owns the task.

Boards selector:
- `j/k`: move between boards
- `enter` or `u`: use board
- `i`: board detail (switches to board if needed)
- `a`: add board
- `r`: rename board
- `ctrl+r` or `F5`: refresh boards/tasks
- `x`: board actions (use/rename/edit description/archive/delete)
- `esc`: back

## 7. Configuration

Editor precedence:
1. `mochi-sticky tui --editor "vim"`
2. `$MOCHI_EDITOR` or `$EDITOR`
3. `.sticky/mochi-sticky.yaml` (`editor` field)
4. Fallback: `nano`

## 8. MCP Usage

Run `mochi-sticky mcp` to start the JSON-RPC server over stdin/stdout. 

Example:
```bash
echo '{"jsonrpc":"2.0","method":"list_tasks","params":{"sort":"priority"},"id":1}' | mochi-sticky mcp
```

See the MCP reference in the built-in wiki: `mochi-sticky wiki view reference/mcp` (or `.sticky/wiki/reference/mcp.md`).

## 9. Contributing

We welcome contributions! Whether it's:
- 🐛 Bug fixes
- ✨ New features
- 📖 Documentation improvements
- 🧪 Test coverage

**Getting Started:**
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

See [MASTER_SPEC.md](MASTER_SPEC.md) for coding standards and architecture details.
Also review:
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [SECURITY.md](SECURITY.md)
- [SUPPORT.md](SUPPORT.md)

## 10. Documentation

- **Installation Guide** — `mochi-sticky wiki view getting-started/install`
- **Quickstart** — `mochi-sticky wiki view getting-started/quickstart`
- **CLI Reference** — `mochi-sticky wiki view user-guide/cli`
- **TUI Guide** — `mochi-sticky wiki view user-guide/tui`
- **MCP Integration** — `mochi-sticky wiki view reference/mcp`
- **Architecture** — `mochi-sticky wiki view development/architecture`
- **Architecture Diagrams** — `mochi-sticky wiki view development/architecture-diagrams`
- **Release Notes Template** — `mochi-sticky wiki view release/release-notes-template`
- **Changelog** — `CHANGELOG.md`
- **Built-in Wiki** — Stored under `.sticky/wiki`

## 11. Philosophy

mochi-sticky follows these core principles:

1. **Data is Code** — Everything lives in your Git repository
2. **Markdown First** — Human-readable, version-controlled files
3. **No External Dependencies** — Single binary, no database required
4. **Developer-Friendly** — CLI, TUI, and API-first design
5. **Agent-Ready** — Built for AI collaboration from the ground up

## 12. License

MIT License - see [LICENSE](LICENSE) for details.

## 13. Community & Support

- 🐛 **Issues** — [Report bugs or request features](https://github.com/mochizero0/mochi-sticky/issues)
- 💬 **Discussions** — [Ask questions and share ideas](https://github.com/mochizero0/mochi-sticky/discussions)
- 📖 **Wiki** — Comprehensive documentation built-in

---

Made with 🍡 for developers who love Markdown and Git.
