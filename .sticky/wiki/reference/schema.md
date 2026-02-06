---
title: Schema Reference
slug: reference/schema
section: Reference
order: 21
tags:
    - schema
    - config
status: published
---
# Schema Reference

Centralized schema definitions for mochi-sticky data files.

## Root Config Schema (v1)

File: `.sticky/mochi-sticky.yaml`

```yaml
storage_root: .sticky
editor: "vim"
config_paths:
  boards: boards/boards.yaml
  adr: adrs/config.yaml
  wiki_index: wiki/_index.yaml
templates:
  root: .sticky/templates
  task: .sticky/templates/task
  board: .sticky/templates/board
  adr: .sticky/templates/adr
  wiki: .sticky/templates/wiki
```

**Fields**:
- `storage_root` *(string, optional)* — storage root directory (default: `.sticky`).
- `editor` *(string, optional)* — editor command used by CLI/TUI.
- `config_paths` *(object, optional)* — locations relative to `storage_root`.
  - `boards` *(string, optional)* — board registry file.
  - `adr` *(string, optional)* — ADR config file.
  - `wiki_index` *(string, optional)* — wiki index file.
- `templates` *(object, optional)* — template paths (see [Configuration](config.md)).

## Board Config Schema (v1)

File: `.sticky/boards/<board-id>/config.yaml`

```yaml
config_version: 1
next_id: 15
columns:
  - title: "Todo"
    key: "todo"
  - title: "Doing"
    key: "doing"
  - title: "Done"
    key: "done"
context:
  scope: "Sprint 42"
  release: "v2.0.0"
```

**Fields**:
- `config_version` *(int, required)* — must be `1` for this schema.
- `next_id` *(int, required)* — next numeric task ID to allocate (defaults to `1` if missing/invalid).
- `columns` *(array, required)* — status columns.
  - `key` *(string, required)* — status identifier used by task frontmatter `status`.
  - `title` *(string, optional)* — display label; defaults to `key` when omitted.
- `context` *(object, optional)* — board-level metadata (see [Configuration](config.md)).

## Board Registry Schema

File: `.sticky/boards/boards.yaml`

```yaml
active: default
boards:
  - id: default
    name: Default
    path: boards/default
    archived: false
    created: "2026-01-29"
```

**Fields**:
- `active` *(string, required)* — currently selected board ID.
- `boards` *(array, required)* — list of all boards.
  - `id` *(string, required)* — directory name (URL-safe).
  - `name` *(string, required)* — display name.
  - `path` *(string, required)* — path under `.sticky/` to the board directory.
  - `archived` *(bool, required)* — whether board is hidden.
  - `created` *(string, required)* — creation date (`YYYY-MM-DD`).

## ADR Config Schema (v1)

File: `.sticky/adrs/config.yaml`

```yaml
config_version: 1
next_id: 13
columns:
  - key: proposed
    title: Proposed
  - key: accepted
    title: Accepted
  - key: rejected
    title: Rejected
  - key: deprecated
    title: Deprecated
  - key: superseded
    title: Superseded
```

**Fields**:
- `config_version` *(int, required)* — must be `1` for this schema.
- `next_id` *(int, required)* — next numeric ADR ID to allocate (defaults to `1` if missing/invalid).
- `columns` *(array, required)* — ADR status columns.
  - `key` *(string, required)* — status identifier used in ADR frontmatter.
  - `title` *(string, optional)* — display label; defaults to `key` when omitted.

## Wiki Index Schema (v1)

File: `.sticky/wiki/_index.yaml`

```yaml
index_version: 1
sections:
  - title: "Architecture"
    slug: "architecture"
    order: 1
    tags: ["core", "design"]
    links:
      depends_on: ["reference"]
      related_to: ["development"]
    pages:
      - "overview"
      - "decisions"
```

**Fields**:
- `index_version` *(int, required)* — must be `1` for this schema.
- `sections` *(array, required)* — section definitions (see [Wiki Schema](wiki-schema.md)).

## Related

- [Configuration](config.md) — Detailed configuration options
- [Boards](boards.md) — Board operations and workflows
- [Wiki Schema](wiki-schema.md) — Wiki storage and index fields
- [ADR](adr.md) — ADR workflow and metadata
