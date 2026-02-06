---
title: PDF Export
slug: export/pdf
section: Export
order: 31
tags:
    - export
    - pdf
status: published
---
# PDF Export

Generate professional PDF documentation from your wiki, perfect for distribution, printing, or archival.

## Prerequisites

### Required: Pandoc

```bash
# macOS
brew install pandoc

# Ubuntu/Debian
sudo apt-get install pandoc texlive-latex-base texlive-latex-extra

# Verify installation
pandoc --version
```

### Required: LaTeX

Pandoc uses LaTeX for PDF generation. Install a TeX distribution for your platform.

## Basic Usage

```bash
# Simple PDF export
mochi-sticky wiki export --format pdf

# With metadata
mochi-sticky wiki export --format pdf \
  --title "Project Documentation" \
  --author "Engineering Team"

# Custom output path
mochi-sticky wiki export --format pdf \
  --output exports/documentation.pdf \
  --title "User Guide"
```

## Options

### --title TITLE
Set document title (appears on cover page).
**Default**: "mochi-sticky Wiki"

### --author AUTHOR
Set document author.

### --output PATH
Specify output file path.
**Default**: `.sticky/wiki/export.pdf`

### --template PATH
Use custom LaTeX template.
**Default**: Built-in template

### --status STATUS
Filter pages by publication status.

## PDF Features

- **Cover Page** — Title, author, date
- **Table of Contents** — Auto-generated with page numbers
- **Page Numbers** — Header/footer navigation
- **Syntax Highlighting** — Code blocks with formatting
- **Hyperlinks** — Internal and external links
- **Images** — Embedded with proper scaling

## Custom Templates

### Using Custom Template

1. Copy default template:
   ```bash
   cp .sticky/templates/wiki/wiki_pdf_template.tex my-template.tex
   ```

2. Customize the LaTeX template

3. Use in export:
   ```bash
   mochi-sticky wiki export --format pdf --template my-template.tex
   ```

### Template Variables

Available variables:
- `$title$` — Document title
- `$author$` — Document author
- `$date$` — Generation date
- `$body$` — Wiki content

## Common Use Cases

### Release Documentation

```bash
mochi-sticky wiki export --format pdf \
  --title "Product v2.0 Documentation" \
  --author "Product Team" \
  --status published \
  --output dist/docs-v2.0.pdf
```

### User Manual

```bash
mochi-sticky wiki export --format pdf \
  --title "User Manual" \
  --author "Support Team" \
  --output manuals/user-guide.pdf
```

## Troubleshooting

### "pandoc: command not found"

**Solution**: Install pandoc (see Prerequisites)

### LaTeX Errors

**Solutions**:
1. Install required LaTeX packages
2. Check for special characters in wiki content
3. Simplify template if using custom one

### Font Issues

```bash
# Install additional fonts
sudo apt-get install texlive-fonts-extra
```

## Best Practices

- **Test locally before CI/CD**

- **Version PDFs** with dates:
```bash
mochi-sticky wiki export --format pdf \
  --output "exports/wiki-$(date +%Y%m%d).pdf"
```

- **Use published status** for production:
```bash
mochi-sticky wiki export --format pdf --status published
```

## Related

- [Markdown Export](markdown.md) — Export to markdown
- [Multi-root Export](multi-root.md) — Combine multiple wikis
- `wiki manifest` — View export order
