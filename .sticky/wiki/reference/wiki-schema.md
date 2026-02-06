---
title: Wiki Schema
slug: reference/wiki-schema
section: Reference
order: 23
tags:
    - wiki
status: published
---
# Wiki Storage and Index Schema

This document defines the on-disk layout and metadata for the mochi-sticky wiki feature.

## Storage Layout
All wiki content lives under `.sticky/wiki/` in the project root.

```
.sticky/wiki/
├── _index.yaml
├── home.md
├── architecture/
│   ├── overview.md
│   └── decisions.md
└── ops/
    ├── runbooks.md
    └── release.md
```

Rules:
- Pages are Markdown files (`.md`).
- Nested folders represent sections.
- Page slugs map to file paths relative to `.sticky/wiki/` (e.g., `architecture/overview`).

## Page Frontmatter
Each page starts with YAML frontmatter that defines metadata used by navigation, filtering, and export.

Fields:
- `title` (string, required): Human-friendly page title.
- `slug` (string, required): Path-like identifier (e.g., `architecture/overview`).
- `section` (string, optional): Logical grouping for display (e.g., `Architecture`).
- `order` (int, optional): Ordering within a section or list.
- `tags` (string list, optional): Used for search and related pages.
- `status` (string, optional): `draft` | `published` | `archived` (defaults to `published`).

Templates:
- Default location: `.sticky/templates/wiki/`
- Use with `mochi-sticky wiki create --template <name>`
- Legacy `.sticky/wiki/templates/` can be migrated by running `mochi-sticky init`, or re-enabled by setting `templates.wiki` in `.sticky/mochi-sticky.yaml`

Example:
```
---
title: "Architecture Overview"
slug: "architecture/overview"
section: "Architecture"
order: 10
tags: ["architecture", "core"]
status: "published"
---
# Architecture Overview
Content here.
```

## Index File: `_index.yaml`
The index defines navigation order and hierarchy. It is optional but recommended for deterministic ordering.

Schema:
- `index_version` (int): schema version (current: `1`).
- `sections`: list of section objects
  - `title` (string): Section title
  - `slug` (string): Section slug (folder name)
  - `order` (int, optional): Section ordering
  - `tags` (string list, optional): Section tags for filtering/grouping
  - `links` (object, optional): Section relationships
    - `depends_on` (string list, optional): Sections this section depends on
    - `related_to` (string list, optional): Sections that are related
  - `pages` (list): Ordered list of page slugs (relative to section)

Example:
```
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
  - title: "Operations"
    slug: "ops"
    order: 2
    pages:
      - "runbooks"
      - "release"
```

## Ordering Rules
- If `_index.yaml` exists, use its order for navigation and export.
- If `_index.yaml` is missing, fall back to alphabetical by `title`.
- If `order` is present on pages/sections, prefer it within the same list.
