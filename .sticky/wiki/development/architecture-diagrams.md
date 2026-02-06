---
title: Architecture Diagrams
slug: development/architecture-diagrams
section: Development
order: 0
tags: []
status: published
---
# Architecture Diagrams

This document contains Mermaid diagrams for quick, shared understanding of the mochi-sticky architecture.

## System Context

```mermaid
flowchart LR
  User[User (CLI/TUI)] -->|commands| App[m🍡ochi-sticky]
  Agent[Agent (MCP client)] -->|JSON-RPC| App
  App -->|reads/writes| FS[(.sticky/ filesystem)]
  App -->|exports| Artifacts[(markdown/pdf exports)]
```

## Components (Internal Packages)

```mermaid
flowchart TB
  CLI[cmd/*] -->|invokes| BoardRepo[internal/board]
  CLI -->|invokes| Storage[internal/storage]
  TUI[internal/tui] -->|reads/writes| BoardRepo
  MCP[internal/mcp] -->|reads/writes| BoardRepo
  MCP --> Wiki[internal/wiki]
  CLI --> Wiki
  Wiki --> FS[(.sticky/wiki)]
  BoardRepo --> FS[(.sticky/boards)]
  ADR[internal/adr] --> FS[(.sticky/adrs)]
  Templates[internal/templates] --> FS[(.sticky/templates)]
```

## Task Write Flow

```mermaid
sequenceDiagram
  participant User
  participant TUI
  participant BoardRepo
  participant FS
  User->>TUI: edit task
  TUI->>BoardRepo: UpdateTask*
  BoardRepo->>FS: write task .md
  FS-->>BoardRepo: write ok
  BoardRepo-->>TUI: updated state
```

## MCP Read/Write Flow

```mermaid
sequenceDiagram
  participant Agent
  participant MCP
  participant BoardRepo
  participant Wiki
  participant FS
  Agent->>MCP: call_tool(list_tasks)
  MCP->>BoardRepo: GetAllTasks
  BoardRepo->>FS: read tasks
  FS-->>BoardRepo: data
  BoardRepo-->>MCP: tasks
  MCP-->>Agent: response
  Agent->>MCP: call_tool(search_wiki)
  MCP->>Wiki: SearchPages
  Wiki->>FS: read wiki files
  FS-->>Wiki: data
  Wiki-->>MCP: results
  MCP-->>Agent: response
```

## Export Flow

```mermaid
flowchart LR
  User[User/Agent] -->|export_wiki| MCP
  MCP --> Wiki
  Wiki -->|render| Export[markdown bytes]
  Export -->|write| FS[(export file)]
  MCP -->|optional| PDF[pandoc]
  PDF --> FS
```
