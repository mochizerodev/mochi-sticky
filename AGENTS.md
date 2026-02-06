# AGENTS.md

This project supports Model Context Protocol (MCP) via the `mochi-sticky mcp` command, enabling AI agents like GitHub Copilot to interact with your tasks, boards, and documentation.

## Capabilities

**37 Tools Available** for:
- ✅ **Tasks**: List, create, update, archive, restore, delete, search by tags/status
- ✅ **Boards**: List, create, rename, set active, archive, delete
- ✅ **Wiki**: List pages, read, write, search, export (markdown/PDF)
- ✅ **Dependencies**: Set task dependencies, list ready tasks
- ✅ **Resources**: Read task/board content, config, registry

## MCP Standard Compliance

The server implements the [Model Context Protocol](https://modelcontextprotocol.io) standard:

- `initialize` - Server initialization with workspace roots support
- `tools/list` - Returns available tools with JSON schemas
- `tools/call` - Invokes tools by name with arguments
- `resources/list` - Returns available resources
- `resources/read` - Reads resources by URI (e.g., `task://T-000001`)
- `shutdown` / `exit` - Graceful shutdown

## VS Code / GitHub Copilot Integration

### Setup

1. **Add to global MCP config** (`~/.config/Code/User/mcp.json`):

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

2. **Reload VS Code** - The mochi-sticky tools will appear in Copilot Chat

3. **Use naturally** - Ask Copilot:
   - "List all tasks in mochi-sticky"
   - "Show me high priority tasks"
   - "Create a new task called 'Fix bug'"
   - "What boards do I have?"

### How It Works

**One Server Per Workspace:**
- VS Code spawns a separate MCP server process for each workspace
- The workspace root is passed via the `initialize` request
- Each server automatically uses that workspace's `.sticky` folder
- No need to specify paths in individual requests

**Example:**
```
Workspace: /home/user/project-a
  → Server uses: /home/user/project-a/.sticky

Workspace: /home/user/project-b  
  → Server uses: /home/user/project-b/.sticky
```

## Connection Modes

- **Persistent**: Server processes multiple requests over stdin/stdout until EOF (default for IDE integrations)
- **One-shot**: Send single request via pipe, server processes and exits (shell scripting)
- Both modes work automatically - client controls lifecycle

## Command Line Options

```bash
mochi-sticky mcp [flags]

Flags:
  --root string      Storage root directory (overrides MOCHI_STICKY_ROOT env)
  --timeout string   Idle timeout (e.g., '30m', '1h') - exits after inactivity
```

Environment variables:
- `MOCHI_STICKY_ROOT` - Default workspace root
- `MOCHI_MCP_TIMEOUT` - Default idle timeout

## Safety

- Destructive actions require `force: true`
- MCP runs over stdin/stdout - do not log to stdout
- Each workspace is isolated - no cross-contamination

## Quick Examples

### Shell Usage

```bash
# List tasks sorted by priority
echo '{"jsonrpc":"2.0","method":"list_tasks","params":{"sort":"priority"},"id":1}' | mochi-sticky mcp

# List all boards
echo '{"jsonrpc":"2.0","method":"list_boards","id":1}' | mochi-sticky mcp

# Create a task
echo '{"jsonrpc":"2.0","method":"create_task","params":{"title":"New task","status":"todo"},"id":1}' | mochi-sticky mcp
```

### Standard MCP Protocol

```bash
# Initialize with workspace
echo '{"jsonrpc":"2.0","method":"initialize","params":{"roots":[{"uri":"file:///path/to/workspace"}]},"id":1}' | mochi-sticky mcp

# List available tools
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | mochi-sticky mcp

# Call a tool
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_tasks","arguments":{}},"id":1}' | mochi-sticky mcp
```

## Troubleshooting

**Tools not appearing in VS Code:**
1. Verify binary is in PATH: `which mochi-sticky`
2. Check Output panel: View → Output → "Model Context Protocol"
3. Reload VS Code window
4. Ensure `.sticky` folder exists: `mochi-sticky init`

**Wrong workspace:**
- Check MCP server is using correct workspace root
- Each VS Code window spawns its own server process
- Server logs show workspace path on startup
