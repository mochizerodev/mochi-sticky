---
title: Multi-root Export
slug: export/multi-root
section: Export
order: 32
tags:
    - export
status: published
---
# Multi-root Export

Combine documentation from multiple mochi-sticky projects into a single export, perfect for mono-repos, multi-service architectures, or consolidated documentation.

## Overview

Multi-root export allows you to:
- Merge wikis from different projects
- Create unified documentation across services
- Combine team-specific wikis
- Generate company-wide documentation

## Basic Usage

```bash
# Export main wiki + one external source
mochi-sticky wiki export --format md \
  --root /path/to/other/project/.sticky/wiki:external \
  --prefix main \
  --output combined.md

# Export with multiple sources
mochi-sticky wiki export --format md \
  --root ../api-service/.sticky/wiki:api \
  --root ../web-app/.sticky/wiki:web \
  --root ../mobile/.sticky/wiki:mobile \
  --prefix platform \
  --output exports/complete-docs.md
```

## Syntax

### --root PATH:PREFIX

Specify external wiki roots with prefixes:

```bash
--root /path/to/wiki:prefix
```

**Components**:
- `PATH` — Absolute or relative path to `.sticky/wiki/` directory
- `:PREFIX` — Namespace for pages from this root

**Examples**:
```bash
--root ../services/auth/.sticky/wiki:auth
--root /home/user/projects/api/.sticky/wiki:api
--root ./external-wiki:ext
```

### --prefix PREFIX

Set prefix for the main wiki (current project):

```bash
--prefix main
```

## How Prefixes Work

### Slug Transformation

Prefixes are prepended to page slugs to avoid conflicts:

```
Original slug: getting-started/install
With prefix "api": api/getting-started/install
```

### Example Structure

```
Main wiki (prefix: main)
  ├── main/home
  ├── main/getting-started/install
  └── main/user-guide/cli

API wiki (prefix: api)
  ├── api/home
  ├── api/endpoints/users
  └── api/authentication

Web wiki (prefix: web)
  ├── web/home
  ├── web/components/button
  └── web/routing
```

### Cross-Referencing

Links between wikis work with prefixed slugs:

```markdown
<!-- In main wiki -->
See [MCP Reference](../reference/mcp.md) for details.

<!-- In API wiki -->
Refer to [CLI Usage](../user-guide/cli.md).
```

## Common Scenarios

### Mono-Repo Documentation

```bash
#!/bin/bash
# Export all service documentation

mochi-sticky wiki export --format md \
  --prefix platform \
  --root services/auth/.sticky/wiki:auth \
  --root services/api/.sticky/wiki:api \
  --root services/web/.sticky/wiki:web \
  --output exports/complete-guide.md
```

### Multi-Team Documentation

```bash
# Combine team wikis
mochi-sticky wiki export --format pdf \
  --prefix engineering \
  --root ../backend-team/.sticky/wiki:backend \
  --root ../frontend-team/.sticky/wiki:frontend \
  --root ../devops-team/.sticky/wiki:devops \
  --title "Engineering Documentation" \
  --output exports/engineering.pdf
```

## Best Practices

- **Use consistent prefixes** across exports:
```bash
# Good
--prefix backend
--root ../frontend/.sticky/wiki:frontend
```

- **Document prefix convention** in README:
```markdown
## Documentation Prefixes
- `main` — Platform core
- `api` — API service
- `web` — Web application
```

- **Validate paths** before export:
```bash
# Check roots exist
for root in ../api/.sticky/wiki ../web/.sticky/wiki; do
  [ -d "$root" ] || echo "Missing: $root"
done
```

- **Use absolute paths** in scripts:
```bash
ROOT_DIR=$(pwd)
mochi-sticky wiki export --format md \
  --root "$ROOT_DIR/../api/.sticky/wiki:api" \
  --prefix main
```

## Troubleshooting

### "Root directory not found"

**Solutions**:
1. Use absolute paths: `--root /full/path/to/.sticky/wiki:prefix`
2. Verify path exists: `ls -la ../other-project/.sticky/wiki`
3. Check relative paths: `readlink -f ../other/.sticky/wiki`

### Duplicate Slugs

**Solution**: Use distinct prefixes for each root:
```bash
--prefix main \
--root ../api/.sticky/wiki:api \
--root ../web/.sticky/wiki:web
```

### Broken Cross-References

Update links to use prefixed slugs:
```markdown
<!-- Before -->
[Reference Docs](../reference/mcp.md)

<!-- After (with prefix) -->
[Reference Docs](../reference/mcp.md)
```

## Related

- [Markdown Export](markdown.md) — Single-wiki markdown export
- [PDF Export](pdf.md) — PDF generation options
- `wiki manifest` — View page ordering
- [Architecture](../development/architecture.md) — System design
