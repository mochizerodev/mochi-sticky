---
title: MCP Usage
slug: user-guide/mcp
section: User Guide
order: 13
tags:
    - mcp
status: published
---
---
title: MCP Integration
slug: user-guide/mcp
section: User Guide
order: 13
tags:
    - mcp
    - api
status: published
---
# MCP Integration

Model Context Protocol (MCP) server for AI agent integration.

## Overview

MCP provides a JSON-RPC 2.0 interface over stdin/stdout for programmatic access to mochi-sticky.

**Use cases**:
- AI agent task management (GitHub Copilot, Claude, etc.)
- Automated workflows
- Integration with other tools
- Batch operations
- Custom tooling

## Starting MCP Server

```bash
mochi-sticky mcp
```

**Connection Model**: The server processes JSON-RPC requests from stdin until EOF or explicit shutdown.

**Modes**:
- **Persistent** (IDEs): Server stays running, handling multiple sequential requests over the same connection
- **One-shot** (Shell): Single request piped in, server exits after response when stdin closes

**Timeout** (optional):
```bash
mochi-sticky mcp --timeout 30m  # Exit after 30 minutes of inactivity
```

**Do NOT** log to stdout while MCP is running.

## MCP Standard Compliance

The server implements the [Model Context Protocol](https://modelcontextprotocol.io) standard:

- `initialize` - Server initialization with workspace roots support
- `tools/list` - Returns available tools with JSON schemas
- `tools/call` - Invokes tools by name with arguments
- `resources/list` - Returns available resources
- `resources/read` - Reads resources by URI (e.g., `task://T-000001`)
- `shutdown` / `exit` - Graceful shutdown

## VS Code & GitHub Copilot Integration

### Global Configuration (Recommended)

Add to `~/.config/Code/User/mcp.json`:

```json
{
  "servers": {
    "mochi-sticky": {
      "type": "stdio",
      "command": "/usr/local/bin/mochi-sticky",
      "args": ["mcp"]
    }
  }
}
```

### How It Works - Workspace Detection

**One Server Per Workspace:**
- VS Code spawns a separate MCP server process for each workspace
- The workspace root is passed via the `initialize` request with `roots` array
- Each server automatically uses that workspace's `.sticky` folder
- No need to specify paths in individual requests

**Example:**
```
Workspace: /home/user/project-a
  → Server uses: /home/user/project-a/.sticky

Workspace: /home/user/project-b
  → Server uses: /home/user/project-b/.sticky
```

### Requirements

1. **mochi-sticky initialized**: Run `mochi-sticky init` in your workspace
2. **`.sticky` directory exists**: Contains boards, tasks, and wiki
3. **GitHub Copilot** with MCP support enabled
4. **Binary in PATH**: Ensure `mochi-sticky` is accessible globally

### Example Prompts

Once configured, you can ask Copilot naturally:

- "List all tasks in mochi-sticky"
- "Show me high priority tasks"
- "Create a new task called 'Fix login bug'"
- "Update task T-000042 to 'doing' status"
- "What boards do I have?"
- "Show me the wiki page about CLI usage"

### Troubleshooting

**Tools not appearing:**
1. Verify binary is in PATH: `which mochi-sticky`
2. Check Output panel: View → Output → "Model Context Protocol"
3. Reload VS Code window
4. Ensure `.sticky` folder exists: `mochi-sticky init`

**Wrong workspace:**
- Each VS Code window spawns its own server process
- Server automatically detects workspace from `initialize` request
- Check MCP server logs for workspace path

## Communication Protocol

### Standard MCP Methods

```bash
# Initialize with workspace
echo '{"jsonrpc":"2.0","method":"initialize","params":{"roots":[{"uri":"file:///path/to/workspace"}]},"id":1}' | mochi-sticky mcp

# List available tools
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | mochi-sticky mcp

# Call a tool
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_tasks","arguments":{"sort":"priority"}},"id":1}' | mochi-sticky mcp

# List resources
echo '{"jsonrpc":"2.0","method":"resources/list","id":1}' | mochi-sticky mcp

# Read a resource
echo '{"jsonrpc":"2.0","method":"resources/read","params":{"uri":"task://T-000001"},"id":1}' | mochi-sticky mcp
```

### Legacy Format (Backward Compatible)

Direct method calls still work:

```json
{
  "jsonrpc": "2.0",
  "method": "list_tasks",
  "params": {
    "sort": "priority"
  },
  "id": 1
}
```

### Response Format

```json
{
  "jsonrpc": "2.0",
  "result": [
    {"id": "T-000001", "title": "Task 1", "status": "todo"}
  ],
  "id": 1
}
```

### Error Format

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32600,
    "message": "Invalid request"
  },
  "id": 1
}
```

## Available Tools (37 Total)

### Task Management
- `list_tasks`, `get_task`, `create_task`
- `update_task_status`, `update_task_priority`, `update_task_title`
- `update_task_tags`, `update_task_content`
- `update_task_dependencies`, `get_task_dependencies`, `list_ready_tasks`
- `archive_task`, `restore_task`, `delete_task`, `list_archived_tasks`

### Board Management
- `list_boards`, `create_board`, `rename_board`
- `set_active_board`, `archive_board`, `delete_board`
- `update_board_description`, `update_board_context`, `get_board_context`

### Wiki Management
- `list_wiki_pages`, `list_wiki_sections`, `read_wiki_page`
- `write_wiki_page`, `update_wiki_section`, `search_wiki`
- `list_wiki_templates`, `create_wiki_from_template`
- `lint_wiki`, `manifest_wiki`, `export_wiki`
- `delete_wiki_page`, `generate_wiki_index`

### Resources
- `task://<id>` - Read full task markdown
- `board://<id>` - Read board description markdown
- `config://active` - Read active board config
- `boards://registry` - Read board registry

## Safety Features

### Force Flag

Destructive operations require `force: true`:

```json
// Delete task
{"method": "delete_task", "params": {
  "id": "T-000042",
  "force": true  // Required
}, "id": 1}
```

### Validation

All inputs validated before execution:
- Task IDs must exist
- Statuses must match board columns
- Slugs must be unique
- Required fields enforced

## Example Usage

### Shell Script

```bash
#!/bin/bash

# List high-priority tasks
echo '{"jsonrpc":"2.0","method":"list_tasks","params":{"sort":"priority","desc":true},"id":1}' | \
  mochi-sticky mcp | jq -r '.result[0:3] | .[] | "\(.priority) - \(.title)"'

# Create task
echo '{"jsonrpc":"2.0","method":"create_task","params":{"title":"Fix bug","priority":1},"id":2}' | \
  mochi-sticky mcp
```

### Python Client

```python
import subprocess
import json

def mcp_call(method, params):
    request = {
        "jsonrpc": "2.0",
        "method": method,
        "params": params,
        "id": 1
    }
    
    proc = subprocess.run(
        ['mochi-sticky', 'mcp'],
        input=json.dumps(request),
        capture_output=True,
        text=True
    )
    
    return json.loads(proc.stdout)

# List tasks using standard protocol
result = mcp_call("list_tasks", {"sort": "priority"})
tasks = result["result"]
for task in tasks[:5]:
    print(f"[{task['priority']}] {task['title']}")
```

## Best Practices

- **Use standard MCP methods** when possible (`tools/call`, `resources/read`)
- **Include request ID**: Helps match responses
- **Handle errors**: Check for `error` field in response
- **Use force flag carefully**: Prevents accidental deletions
- **Validate inputs**: Check data before sending
- **One workspace = one server**: Let MCP handle workspace detection

## Error Codes

Standard JSON-RPC error codes:

- `-32700` — Parse error (invalid JSON)
- `-32600` — Invalid request (missing fields)
- `-32601` — Method not found
- `-32602` — Invalid params
- `-32603` — Internal error

## Related

- [MCP Reference](../reference/mcp.md) — Complete API documentation
- [CLI Usage](cli.md) — Command-line interface
- [Tasks Reference](../reference/tasks.md) — Task structure
- [Boards Reference](../reference/boards.md) — Board management

For complete setup instructions, see [AGENTS.md](../../AGENTS.md) in the repository root.
