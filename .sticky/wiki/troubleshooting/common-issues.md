---
title: Troubleshooting
slug: troubleshooting/common-issues
section: Troubleshooting
order: 50
tags:
    - troubleshooting
status: published
---
# Troubleshooting

Common issues and solutions for mochi-sticky.

## Installation Issues

### Build Fails

**Problem**: `go build` returns errors

**Solutions**:

1. **Check Go version**:
   ```bash
   go version  # Requires Go 1.21+
   ```

2. **Update dependencies**:
   ```bash
   go mod tidy
   go mod download
   ```

3. **Clear module cache**:
   ```bash
   go clean -modcache
   go build -o mochi-sticky
   ```

### Binary Not Found After Install

**Problem**: `mochi-sticky: command not found`

**Solutions**:

1. **Check installation path**:
   ```bash
   which mochi-sticky
   ls -la /usr/local/bin/mochi-sticky
   ```

2. **Verify PATH includes installation directory**:
   ```bash
   echo $PATH | grep /usr/local/bin
   ```

3. **Add to PATH** if missing:
   ```bash
   export PATH="/usr/local/bin:$PATH"
   # Add to ~/.bashrc or ~/.zshrc for persistence
   ```

4. **Reinstall with correct permissions**:
   ```bash
   sudo cp -f mochi-sticky /usr/local/bin
   sudo chmod +x /usr/local/bin/mochi-sticky
   ```

## Storage & Initialization

### Storage Root Not Found

**Error**: `storage root does not exist`

**Solutions**:

1. **Initialize storage**:
   ```bash
   mochi-sticky init
   ```

2. **Check storage location**:
   ```bash
   # Default location
   ls -la .sticky/
   
   # Or check configuration
   cat .sticky/mochi-sticky.yaml
   ```

3. **Specify custom storage**:
   ```bash
   mochi-sticky --storage /custom/path init
   mochi-sticky --storage /custom/path task list
   ```

4. **Fix permissions**:
   ```bash
   chmod -R u+rw .sticky/
   ```

### Invalid Configuration

**Error**: `failed to parse config`

**Solutions**:

1. **Validate YAML syntax**:
   ```bash
   # Use a YAML validator
   cat .sticky/boards/default/config.yaml | python3 -c "import yaml, sys; yaml.safe_load(sys.stdin)"
   ```

2. **Reset to defaults**:
   ```bash
   mochi-sticky init --force
   ```

3. **Check for required fields**:
   ```yaml
   # Board config must have:
   board:
     columns:
       - name: "To Do"
         key: "todo"
   next_id: 1
   ```

### Hydration Fails

**Error**: `hydration failed`

**Solutions**:

1. **Run hydration with details**:
   ```bash
   mochi-sticky hydrate --json --pretty
   ```

2. **Check board registry**:
   ```bash
   cat .sticky/boards/boards.yaml
   ```

3. **Verify active board exists**:
   ```bash
   ls .sticky/boards/default/
   ```

4. **Fix manually or reinitialize**:
   ```bash
   mochi-sticky init --force
   ```

## Task Management

### Task Not Found

**Error**: `task not found: T-000042`

**Solutions**:

1. **Check if task is in different board**:
   ```bash
   mochi-sticky board list
   mochi-sticky board use other-board
   mochi-sticky task show T-000042
   ```

2. **Check if task is archived**:
   ```bash
   mochi-sticky task archive list
   mochi-sticky task archive restore T-000042
   ```

3. **List all tasks to find it**:
   ```bash
   mochi-sticky task list
   ```

4. **Verify task ID format**:
   - Must be `T-NNNNNN` (6 digits)
   - Check for typos

### Invalid Status Error

**Error**: `invalid status: xyz`

**Solutions**:

1. **List valid statuses for current board**:
   ```bash
   mochi-sticky task statuses
   ```

2. **Check board configuration**:
   ```bash
   cat .sticky/boards/default/config.yaml
   ```

3. **Update task with valid status**:
   ```bash
   mochi-sticky task move T-000042 todo
   ```

4. **Add missing column to board config**:
   ```yaml
   board:
     columns:
       - name: "Display Name"
         key: "xyz"  # Add this status
   ```

### Cannot Archive Task

**Error**: `task cannot be archived`

**Solutions**:

1. **Use force flag for active tasks**:
   ```bash
   mochi-sticky task archive task T-000042 --force
   ```

2. **Move to done before archiving**:
   ```bash
   mochi-sticky task move T-000042 done
   mochi-sticky task archive task T-000042
   ```

### Circular Dependencies

**Problem**: Task depends on itself

**Solutions**:

1. **View dependency chain**:
   ```bash
   mochi-sticky task deps T-000042
   ```

2. **Clear dependencies**:
   ```bash
   # Edit task file and remove circular reference
   mochi-sticky task show T-000042
   # Or set dependencies to empty
   mochi-sticky task deps T-000042 --set ""
   ```

3. **Redesign dependency structure**:
   - Break into smaller tasks
   - Remove circular references

## Board Management

### Wrong Board Active

**Problem**: Commands affecting unexpected board

**Solutions**:

1. **Check active board**:
   ```bash
   mochi-sticky board show
   ```

2. **Switch to correct board**:
   ```bash
   mochi-sticky board use project-alpha
   ```

3. **Verify board selection**:
   ```bash
   mochi-sticky board list
   ```

4. **Set in environment** for consistency:
   ```bash
   export MOCHI_STICKY_BOARD=project-alpha
   ```

### Board Not Found

**Error**: `board not found: xyz`

**Solutions**:

1. **List available boards**:
   ```bash
   mochi-sticky board list
   mochi-sticky board list --include-archived
   ```

2. **Check board ID** (not display name):
   ```bash
   # Use ID from registry
   cat .sticky/boards/boards.yaml
   ```

3. **Create board if missing**:
   ```bash
   mochi-sticky board add "New Board"
   ```

### Cannot Delete Board

**Error**: Deletion requires `--force`

**Solutions**:

1. **Archive first** (reversible):
   ```bash
   mochi-sticky board archive old-board
   ```

2. **Force delete** (permanent):
   ```bash
   mochi-sticky board delete old-board --force
   ```

3. **Export data before deletion**:
   ```bash
   # Switch to board
   mochi-sticky board use old-board
   
   # Export tasks
   mochi-sticky task list > board-backup.txt
   
   # Then delete
   mochi-sticky board delete old-board --force
   ```

## TUI Issues

### TUI Display Garbled

**Problem**: Incorrect layout or garbled characters

**Solutions**:

1. **Increase terminal size**:
   - Minimum: 80 columns × 24 rows
   - Recommended: 120 columns × 40 rows

2. **Check terminal emulator**:
   - Use modern terminal with Unicode support
   - Test: `echo "┌─┬─┐│├─┼─┤│└─┴─┘"`

3. **Verify TERM variable**:
   ```bash
   echo $TERM  # Should be xterm-256color or similar
   export TERM=xterm-256color
   ```

4. **Try different terminal**:
   - iTerm2 (macOS)
   - Windows Terminal (Windows)
   - Alacritty, kitty (Linux)

### TUI Freezes

**Problem**: TUI becomes unresponsive

**Solutions**:

1. **Force quit**: Press `Ctrl+C`

2. **Check for blocking operations**:
   - Large number of tasks
   - Slow file system

3. **Restart TUI**:
   ```bash
   mochi-sticky tui
   ```

### Cannot Edit Tasks in TUI

**Problem**: Editor doesn't open

**Solutions**:

1. **Set editor explicitly**:
   ```bash
   mochi-sticky tui --set-editor "nano"
   ```

2. **Check editor is installed**:
   ```bash
   which nano vim code
   ```

3. **Use full path**:
   ```bash
   mochi-sticky tui --set-editor "/usr/bin/nano"
   ```

4. **Set environment variable**:
   ```bash
   export MOCHI_EDITOR="vim"
   export EDITOR="vim"
   ```

## Wiki Issues

### Page Not Found

**Error**: `page not found: xyz`

**Solutions**:

1. **Check slug format**:
   ```bash
   mochi-sticky wiki list
   # Use exact slug: section/page-name
   ```

2. **Verify page exists**:
   ```bash
   ls .sticky/wiki/user-guide/
   ```

3. **Check manifest**:
   ```bash
   cat .sticky/wiki/manifest.yaml
   ```

4. **Regenerate index**:
   ```bash
   mochi-sticky wiki index generate
   ```

### Broken Links in Export

**Problem**: Links don't work in PDF/Markdown export

**Solutions**:

1. **Use relative paths**:
   ```markdown
   [Tasks](../reference/tasks.md)  ✅
   [Tasks](../reference/tasks.md)
   ```

2. **Avoid absolute URLs** for internal pages

3. **Run lint before export**:
   ```bash
   mochi-sticky wiki lint
   ```

### PDF Export Fails

**Error**: `pandoc not found` or conversion error

**Solutions**:

1. **Install Pandoc**:
   ```bash
   # macOS
   brew install pandoc
   
   # Ubuntu/Debian
   sudo apt-get install pandoc texlive
   
   # Windows
   choco install pandoc
   ```

2. **Verify installation**:
   ```bash
   pandoc --version
   ```

3. **Check LaTeX** (for PDF):
   ```bash
   which pdflatex
   ```

4. **Use simpler template**:
   ```bash
   mochi-sticky wiki export pdf --output doc.pdf --no-template
   ```

### Duplicate Slugs

**Error**: `duplicate slug: xyz`

**Solutions**:

1. **Find duplicate pages**:
   ```bash
   mochi-sticky wiki lint
   ```

2. **Rename one page**:
   ```bash
   mochi-sticky wiki edit problematic-page
   # Change slug in frontmatter
   ```

3. **Check manifest for duplicates**:
   ```bash
   cat .sticky/wiki/manifest.yaml | grep -c "slug: user-guide/page"
   ```

## MCP Integration

### No Response from MCP

**Problem**: MCP doesn't return anything

**Solutions**:

1. **Check JSON syntax**:
   ```bash
   cat request.json | jq .
   ```

2. **Verify required fields**:
   ```json
   {
     "jsonrpc": "2.0",
     "method": "list_tasks",
     "params": {},
     "id": 1
   }
   ```

3. **Check for stdout pollution**:
   - MCP reads from stdin, writes to stdout
   - Any debug logging will break response

4. **Test with simple request**:
   ```bash
   echo '{"jsonrpc":"2.0","method":"list_tasks","params":{},"id":1}' | mochi-sticky mcp
   ```

### MCP Invalid Request

**Error**: `Invalid request`

**Solutions**:

1. **Include all required fields**:
   - `jsonrpc: "2.0"`
   - `method: "method_name"`
   - `params: {}` (object, not array)
   - `id: 1`

2. **Check method name spelling**:
   ```bash
   # Correct
   "method": "list_tasks"
   
   # Wrong
   "method": "listTasks"
   "method": "list-tasks"
   ```

3. **Ensure params is object**:
   ```json
   "params": {}        ✅
   "params": []        ❌
   "params": null      ❌
   ```

### Wrong Board in MCP Calls

**Problem**: MCP uses unexpected board

**Solutions**:

1. **Check active board**:
   ```json
   {"method": "read_registry", "params": {}, "id": 1}
   ```

2. **Set active board first**:
   ```json
   {"method": "set_active_board", "params": {
     "board_id": "project-alpha"
   }, "id": 1}
   ```

3. **Ensure consistent board** across requests

## ADR Issues

### Invalid ADR Status

**Error**: `invalid status: xyz`

**Solutions**:

1. **List valid statuses**:
   ```bash
   mochi-sticky adr statuses
   ```

2. **Check ADR config**:
   ```bash
   cat .sticky/adrs/config.yaml
   ```

3. **Use standard statuses**:
   - `proposed`
   - `accepted`
   - `deprecated`
   - `superseded`

### Broken Supersession Chain

**Problem**: ADR references non-existent ADR

**Solutions**:

1. **Run lint**:
   ```bash
   mochi-sticky adr lint
   ```

2. **Fix frontmatter**:
   ```bash
   mochi-sticky adr edit 4
   # Remove invalid ID from supersedes array
   ```

3. **Verify referenced ADRs exist**:
   ```bash
   mochi-sticky adr list
   ```

## Performance Issues

### Slow Task Listing

**Problem**: `task list` takes too long

**Solutions**:

1. **Archive old tasks**:
   ```bash
   mochi-sticky task archive before 2025-12-31
   ```

2. **Use filters** to reduce output:
   ```bash
   mochi-sticky task list --status todo
   mochi-sticky task list --from 2026-01-01
   ```

3. **Check file system performance**:
   - SSD recommended over HDD
   - Avoid network drives for .sticky/

### Large Storage Size

**Problem**: `.sticky/` directory too large

**Solutions**:

1. **Archive completed work**:
   ```bash
   mochi-sticky task archive before 2025-12-31
   ```

2. **Delete old archived tasks**:
   ```bash
   mochi-sticky task archive delete T-000042 --force
   ```

3. **Archive old boards**:
   ```bash
   mochi-sticky board archive old-project
   ```

4. **Clean up wiki exports**:
   ```bash
   rm .sticky/wiki/exports/*.pdf
   ```

## General Troubleshooting Steps

### Debug Mode

Enable verbose logging:

```bash
# Set log level
export MOCHI_LOG_LEVEL=debug

# Run command
mochi-sticky task list
```

### Backup Before Major Changes

Always backup your data:

```bash
# Create backup
cp -r .sticky .sticky.backup.$(date +%Y%m%d)

# Restore if needed
rm -rf .sticky
cp -r .sticky.backup.20260204 .sticky
```

### Validate Installation

Check system health:

```bash
# Validate storage
mochi-sticky hydrate --json --pretty

# Test basic operations
mochi-sticky task add "Test task"
mochi-sticky task list
mochi-sticky task delete T-000001 --force
```

### Getting Help

If issues persist:

1. **Check documentation**:
   - [CLI Usage](../user-guide/cli.md)
   - [Configuration](../reference/config.md)
   - [Tasks Reference](../reference/tasks.md)

2. **Run diagnostics**:
   ```bash
   mochi-sticky hydrate --json --pretty
   mochi-sticky --version
   go version
   ```

3. **Check GitHub issues**: Look for similar problems

4. **Create minimal reproduction**:
   - Fresh initialization
   - Minimal steps to reproduce
   - Include error messages

## Related

- [Installation](../getting-started/install.md) — Setup guide
- [CLI Usage](../user-guide/cli.md) — Command reference
- [Configuration](../reference/config.md) — Configuration files
- [MCP Usage](../user-guide/mcp.md) — API integration
