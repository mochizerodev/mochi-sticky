---
title: MCP Reference
slug: reference/mcp
section: Reference
order: 0
tags: []
status: published
---
# MCP Guide

This document explains how to run the MCP server for `mochi-sticky`, and how to integrate it with VS Code (Copilot) and Codex. MCP runs over stdin/stdout using JSON-RPC 2.0.

## 1. What is MCP here?

`mochi-sticky mcp` starts a JSON-RPC server over stdin/stdout. It exposes tools and resources so an agent can manage boards and tasks.

Key points:
- No TCP port is used (stdin/stdout only).
- Destructive actions require `force: true`.
- Task responses include `board_id` (active board if not provided).

## 2. Run the MCP server manually

```bash
mochi-sticky mcp
```

Test with a one-shot request:
```bash
echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | mochi-sticky mcp
```

List tasks (active board):
```bash
echo '{"jsonrpc":"2.0","method":"list_tasks","params":{"sort":"priority"},"id":2}' | mochi-sticky mcp
```

List tasks for a specific board:
```bash
echo '{"jsonrpc":"2.0","method":"list_tasks","params":{"board_id":"board-2","sort":"priority"},"id":3}' | mochi-sticky mcp
```

## 3. Tool and resource overview

Tools (examples):
- `list_tasks`, `get_task`, `create_task`
- `update_task_status`, `update_task_priority`, `update_task_title`, `update_task_tags`, `update_task_content`
- `archive_task`, `restore_task`, `delete_task`, `list_archived_tasks`
- `list_boards`, `create_board`, `rename_board`, `set_active_board`, `archive_board`, `delete_board`, `update_board_description`
- `list_wiki_pages`, `read_wiki_page`, `write_wiki_page`, `search_wiki`
- `list_wiki_sections`, `update_wiki_section`
- `list_wiki_templates`, `create_wiki_from_template`, `lint_wiki`, `manifest_wiki`, `export_wiki`
- `delete_wiki_page`, `generate_wiki_index`

Resources:
- `read_task` (task markdown)
- `read_board` (board description markdown)
- `read_config` (board config)
- `read_boards` (board registry)

Destructive tools require:
```json
{"force": true}
```

## 4. VS Code integration (Copilot)

Copilot’s agent tooling can connect to MCP processes by running the command and using stdin/stdout. The configuration can differ by extension version, but the pattern is:

- Provide a command that launches `mochi-sticky mcp`
- Register it as a tool provider
- Use prompts that instruct Copilot to call tools

Example prompt:
```
Use the MCP tools from mochi-sticky to list tasks in the active board, sorted by priority. Then summarize the top 3 items.
```

Another prompt:
```
Create a task "Fix login bug" with priority 1 and tags "backend,auth" using MCP.
```

## 5. VS Code integration (Codex)

Codex connects to MCP over stdin/stdout by spawning the process. Configure the MCP command for the workspace.

Example prompt:
```
List tasks from board "board-2" sorted by priority, then move the top task to status "doing".
```

Example prompt (with delete protection):
```
Archive task T-000123 using MCP (remember to pass force: true).
```

## 6. Options and patterns

Common usage patterns:
- Board scoped calls: pass `board_id` when you need a specific board.
- Safe actions: use list/get first, then update.
- Destructive actions: always set `force: true`.

Example request body (update status):
```json
{"jsonrpc":"2.0","method":"update_task_status","params":{"id":"T-000123","status":"doing"},"id":10}
```

Example request body (update priority):
```json
{"jsonrpc":"2.0","method":"update_task_priority","params":{"id":"T-000123","priority":1},"id":11}
```

Example request body (list wiki pages):
```json
{"jsonrpc":"2.0","method":"list_wiki_pages","params":{"status":"published","include_templates":false},"id":12}
```

Example request body (list wiki sections):
```json
{"jsonrpc":"2.0","method":"list_wiki_sections","params":{"tags":["platform"],"link_type":"depends_on"},"id":12}
```

Example request body (read wiki page):
```json
{"jsonrpc":"2.0","method":"read_wiki_page","params":{"slug":"architecture/overview"},"id":13}
```

Example request body (write wiki page):
```json
{"jsonrpc":"2.0","method":"write_wiki_page","params":{"slug":"architecture/overview","title":"Architecture Overview","content":"# Architecture\\n..."},"id":14}
```

Example request body (update wiki section):
```json
{"jsonrpc":"2.0","method":"update_wiki_section","params":{"slug":"architecture","tags":["core"],"links":{"depends_on":["reference"]}},"id":14}
```

Example request body (search wiki):
```json
{"jsonrpc":"2.0","method":"search_wiki","params":{"query":"database","case_insensitive":true,"include_templates":false,"status":"published"},"id":15}
```

Example request body (list wiki templates):
```json
{"jsonrpc":"2.0","method":"list_wiki_templates","params":{},"id":16}
```

Example request body (create wiki from template):
```json
{"jsonrpc":"2.0","method":"create_wiki_from_template","params":{"title":"Runbook: Cache","slug":"runbooks/cache","template":"runbook"},"id":17}
```

Example request body (lint wiki):
```json
{"jsonrpc":"2.0","method":"lint_wiki","params":{},"id":18}
```

Example request body (manifest wiki):
```json
{"jsonrpc":"2.0","method":"manifest_wiki","params":{},"id":19}
```

Example request body (export wiki):
```json
{"jsonrpc":"2.0","method":"export_wiki","params":{"format":"md","output":".sticky/wiki/export.md","roots":["../other/.sticky/wiki:ext"],"prefix":"main"},"id":20}
```

Example request body (delete wiki page):
```json
{"jsonrpc":"2.0","method":"delete_wiki_page","params":{"slug":"architecture/overview","update_index":true},"id":21}
```

Example request body (generate wiki index):
```json
{"jsonrpc":"2.0","method":"generate_wiki_index","params":{"include_templates":false,"write":true},"id":22}
```

## 7. Troubleshooting

- If responses are missing: ensure only JSON is written to stdout (no debug logs).
- If a destructive call fails: check that `force: true` is included.
- If the wrong board is used: pass `board_id` explicitly.
