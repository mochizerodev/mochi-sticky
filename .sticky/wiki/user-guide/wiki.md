---
title: Wiki Basics
slug: user-guide/wiki
section: User Guide
order: 12
tags:
    - wiki
status: published
---
---
title: Wiki Usage
slug: user-guide/wiki
section: User Guide
order: 12
tags:
    - wiki
status: published
---
# Wiki Usage

Markdown-based documentation system with YAML frontmatter.

## Overview

Wiki pages are stored in `.sticky/wiki/` and organized into sections.

**Features**:
- Markdown with YAML metadata
- Section-based organization
- Full-text search
- Export to Markdown/PDF
- Link validation
- Tag-based categorization

## File Structure

```
.sticky/wiki/
├── manifest.yaml           # Page index
├── home.md                # Landing page
├── getting-started/
│   ├── install.md
│   └── quickstart.md
├── user-guide/
│   ├── cli.md
│   └── tui.md
└── reference/
    ├── config.md
    └── tasks.md
```

## Page Structure

```markdown
---
title: Page Title
slug: section/page-name
section: Section Name
order: 10
tags:
  - tag1
  - tag2
status: published
---

# Page Title

Content in Markdown format.

## Sections

More content here.
```

See [Wiki Schema](../reference/wiki-schema.md) for field details.

## Common Commands

### List Pages

```bash
# All pages
mochi-sticky wiki list

# By section
mochi-sticky wiki list --section "User Guide"

# By tag
mochi-sticky wiki list --tag tutorial

# By status
mochi-sticky wiki list --status draft
```

### Create Page

```bash
mochi-sticky wiki create \
  --title "New Guide" \
  --slug user-guide/new-guide \
  --section "User Guide" \
  --order 40
```

**Opens in editor** for content entry.

### View Page

```bash
# View in terminal
mochi-sticky wiki view getting-started/install

# View with metadata
mochi-sticky wiki view getting-started/install --metadata
```

### Edit Page

```bash
mochi-sticky wiki edit user-guide/cli
mochi-sticky wiki edit user-guide/cli --editor vim
```

### Delete Page

```bash
mochi-sticky wiki delete old-page --force
```

## Search

Full-text search across all wiki pages:

```bash
# Search content
mochi-sticky wiki search "authentication"

# Search in specific section
mochi-sticky wiki search "API" --section Reference

# Search with context
mochi-sticky wiki search "config" --context 3
```

**Output**:
```
user-guide/cli.md:42
  --tags <tag1,tag2> — Comma-separated tags
  --status <key> — Initial status (default: todo)

reference/config.md:15
  Storage root configuration via .sticky/mochi-sticky.yaml
```

## Section Management

Common sections:

- **Getting Started** — Installation, quickstart
- **User Guide** — How-to guides
- **Reference** — Technical documentation
- **Export** — Export guides
- **Development** — Contribution docs

### Create Section

Just add pages with new section name:

```bash
mochi-sticky wiki create \
  --title "First Page" \
  --slug new-section/first-page \
  --section "New Section"
```

## Index Generation

Generate table of contents:

```bash
# Generate index.md
mochi-sticky wiki index generate

# Custom template
mochi-sticky wiki index generate --template /path/to/template.md
```

**Output**: `.sticky/wiki/index.md` with links to all pages.

## Export

### Markdown Export

Single file with all pages:

```bash
mochi-sticky wiki export markdown --output docs.md
```

**Options**:
```bash
--sections "User Guide,Reference"  # Specific sections
--exclude-sections "Development"   # Exclude sections
--flatten                         # Remove section hierarchy
```

See [Markdown Export](../export/markdown.md) for details.

### PDF Export

Requires Pandoc:

```bash
mochi-sticky wiki export pdf --output documentation.pdf
```

**Options**:
```bash
--template /path/to/template.tex  # Custom LaTeX template
--toc                            # Include table of contents
--sections "User Guide"          # Specific sections
```

See [PDF Export](../export/pdf.md) for details.

### Multi-Root Export

Merge wikis from multiple projects:

```bash
mochi-sticky wiki export multi-root \
  --roots /project1/.sticky/wiki,/project2/.sticky/wiki \
  --output merged.md
```

See [Multi-Root Export](../export/multi-root.md) for details.

## Validation

Check wiki health:

```bash
# Lint all pages
mochi-sticky wiki lint

# Fix issues automatically
mochi-sticky wiki lint --fix
```

**Checks**:
- Missing required fields
- Invalid slug format
- Broken internal links
- Duplicate slugs
- Pages not in manifest
- Invalid section references

## Best Practices

- **Consistent slugs**: Use `section/page-name` format

- **Order with gaps**: Use increments of 10 (10, 20, 30)

- **Descriptive titles**: Clear and concise

- **Tag appropriately**: Use for cross-referencing

- **Update manifest**: Run `wiki index generate` after changes

- **Version control**: Commit wiki alongside code

## Common Workflows

### Add Documentation

```bash
# Create page
mochi-sticky wiki create \
  --title "API Authentication" \
  --slug user-guide/auth \
  --section "User Guide" \
  --order 50

# Write content (opens editor)

# Verify
mochi-sticky wiki view user-guide/auth

# Update manifest
mochi-sticky wiki index generate
```

### Update Existing Page

```bash
# Edit
mochi-sticky wiki edit user-guide/cli

# Verify changes
mochi-sticky wiki view user-guide/cli

# Search for references
mochi-sticky wiki search "cli"
```

### Publish Documentation

```bash
# Lint for errors
mochi-sticky wiki lint

# Export to PDF
mochi-sticky wiki export pdf --output user-guide.pdf

# Export to Markdown
mochi-sticky wiki export markdown --output docs.md
```

## Troubleshooting

### Page Not Found

**Error**: "page not found: xyz"

**Solutions**:
1. Check slug format: `mochi-sticky wiki list`
2. Verify page exists: `ls .sticky/wiki/`
3. Check manifest: `cat .sticky/wiki/manifest.yaml`

### Broken Links

**Problem**: Links don't work in exports

**Solution**: Use relative paths:
```markdown
[Tasks](../reference/tasks.md)  -
[Tasks](../reference/tasks.md)
```

### Duplicate Slugs

**Error**: "duplicate slug: user-guide/cli"

**Solution**: Each slug must be unique. Rename one page.

## Related

- [Markdown Export](../export/markdown.md) — Export to single file
- [PDF Export](../export/pdf.md) — Generate PDFs
- [Wiki Schema](../reference/wiki-schema.md) — Frontmatter reference
