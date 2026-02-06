---
title: Markdown Export
slug: export/markdown
section: Export
order: 30
tags:
    - export
status: published
---
# Markdown Export

Export your entire wiki to a single Markdown file, perfect for creating consolidated documentation, README files, or archival purposes.

## Basic Usage

```bash
# Export to default location
mochi-sticky wiki export --format md

# Export to specific file
mochi-sticky wiki export --format md --output exports/wiki.md

# Export with custom prefix
mochi-sticky wiki export --format md --prefix project --output wiki.md
```

## How It Works

1. **Collection**: Gathers all published wiki pages
2. **Ordering**: Sorts by manifest order (or alphabetically if no manifest)
3. **Combining**: Merges pages into a single document with separators
4. **Output**: Writes to specified file path

## Output Structure

The exported file contains:
- Table of contents (auto-generated from page hierarchy)
- Page separators with titles
- Full markdown content from each page
- Preserved code blocks and formatting
- Internal links adjusted for single-file format

## Common Use Cases

### Creating Distribution Docs

```bash
# Export for distribution
mochi-sticky wiki export --format md \
  --status published \
  --output dist/DOCUMENTATION.md
```

### Archiving Documentation

```bash
# Archive with timestamp
mochi-sticky wiki export --format md \
  --output "archive/wiki-$(date +%Y%m%d).md"
```

## Options

### --output PATH
Specify output file path.
**Default**: `.sticky/wiki/export.md`

### --prefix PREFIX
Add prefix to page slugs.

### --include-templates
Include template pages in export.

### --status STATUS
Filter pages by status (`published`, `draft`, `archived`).

## Best Practices

- **Generate manifest before export**:
```bash
mochi-sticky wiki index --write
mochi-sticky wiki export --format md
```

- **Use version control** for exported files

- **Keep exports separate** from source wiki files

## Related Commands

- [PDF Export](pdf.md) — Export to PDF format
- [Multi-root Export](multi-root.md) — Combine multiple wikis
- `wiki manifest` — View export order
- `wiki index` — Generate manifest
