---
title: Installation
slug: getting-started/install
section: Getting Started
order: 1
tags:
    - install
status: published
---
# Installation

This guide covers how to build and install mochi-sticky from source.

## Prerequisites

- **Go 1.23 or later** — [Download Go](https://go.dev/dl/)
- **Git** — For cloning the repository
- A Unix-like environment (Linux, macOS, WSL on Windows)

## Build from Source

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/mochi-sticky.git
cd mochi-sticky
```

### 2. Build the Binary

```bash
go build -o mochi-sticky
```

This creates the `mochi-sticky` executable in the current directory.

### 3. Install Globally (Optional)

To use `mochi-sticky` from anywhere:

**Linux/macOS:**
```bash
sudo cp mochi-sticky /usr/local/bin/
# Or without sudo:
mkdir -p ~/bin
cp mochi-sticky ~/bin/
export PATH="$HOME/bin:$PATH"  # Add to ~/.bashrc or ~/.zshrc
```

**Windows (WSL):**
```bash
sudo cp mochi-sticky /usr/local/bin/
```

### 4. Verify Installation

```bash
mochi-sticky --version
mochi-sticky --help
```

## Initialize Your First Project

Navigate to your project directory and initialize mochi-sticky:

```bash
cd /path/to/your/project
mochi-sticky init
```

This creates:
- `.sticky/` — Data directory for boards, tasks, and configuration
- `.sticky/boards/boards.yaml` — Board registry
- `.sticky/boards/default/` — Default board with initial configuration
- `.sticky/templates/` — Default templates (task, board, ADR, wiki)
- `.sticky/wiki/` — Wiki documentation root
- `.sticky/adrs/` — Architecture Decision Records

**Note**: `mochi-sticky tui --set-editor "code --wait"` writes the editor preference to `.sticky/mochi-sticky.yaml` (`editor`).

## Storage Root Configuration

By default, mochi-sticky uses `.sticky/` in your current directory. You can customize this:

### Method 1: Configuration File

Create `.sticky/mochi-sticky.yaml`:

```yaml
storage_root: .sticky
editor: "code --wait"
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

If you already have a legacy `mochi-sticky.yaml` in the project root, it is still honored for `storage_root`.

### Method 2: Environment Variable

```bash
export MOCHI_STICKY_STORAGE=/path/to/custom/storage
mochi-sticky init
```

### Method 3: CLI Flag

```bash
mochi-sticky --storage /path/to/custom/storage init
```

**Priority order:** CLI flag > Environment variable > `.sticky/mochi-sticky.yaml` > legacy root `mochi-sticky.yaml` > Default (`.sticky/`)

## Validate Installation

Check that everything is configured correctly:

```bash
mochi-sticky hydrate
```

This outputs:
- Storage root location
- Board registry path
- Active board
- Template directories
- Configuration validity

For JSON output (useful for automation):
```bash
mochi-sticky hydrate --json --pretty
```

## Quick Start

Once installed, create your first task:

```bash
mochi-sticky task add "My first task" --tags onboarding --priority 1
mochi-sticky task list
```

Or launch the interactive TUI:

```bash
mochi-sticky tui
```

## Troubleshooting

### "command not found: mochi-sticky"

Ensure the binary is in your `$PATH`:
```bash
echo $PATH
which mochi-sticky
```

### "permission denied"

Make the binary executable:
```bash
chmod +x mochi-sticky
```

### Build Errors

Ensure you have Go 1.21+ installed:
```bash
go version
```

Update dependencies:
```bash
go mod tidy
go mod download
```

## Next Steps

- [Quickstart Guide](../getting-started/quickstart.md) — Create your first task
- [CLI Usage](../user-guide/cli.md) — Learn essential commands
- [TUI Usage](../user-guide/tui.md) — Explore the interactive interface
- [Configuration](../reference/config.md) — Customize your setup
