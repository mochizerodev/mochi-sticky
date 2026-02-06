---
title: Wiki Templates
slug: reference/templates
section: Reference
order: 24
tags:
    - wiki
    - templates
status: published
---
---
title: Templates
slug: reference/templates
section: Reference
order: 23
tags:
    - templates
status: published
---
# Templates

mochi-sticky uses templates to generate new files with consistent structure.

## Template Locations

Default template directory: `.sticky/templates/`

```
.sticky/templates/
├── task.md          # Task template
├── board.yaml       # Board configuration template
├── adr.md          # ADR template
└── wiki/
    ├── page.md             # Wiki page template
    └── wiki_pdf_template.tex # LaTeX template for PDF export
```

## Task Template

Located at `.sticky/templates/task.md`:

```markdown
---
id: "{{.ID}}"
uid: "{{.UID}}"
title: "{{.Title}}"
status: "{{.Status}}"
priority: 2
tags: []
created: {{.Created}}
dependencies: []
---

# Task Description

Describe the task here.

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2

## Notes

Add implementation notes, design decisions, or links to resources.
```

**Variables**:
- `{{.ID}}` — Generated task ID (T-000042)
- `{{.UID}}` — Generated UUID
- `{{.Title}}` — Title from command
- `{{.Status}}` — Initial status (usually "todo")
- `{{.Created}}` — Current date (YYYY-MM-DD)

## Board Template

Located at `.sticky/templates/board.yaml`:

```yaml
board:
  columns:
    - name: "To Do"
      key: "todo"
    - name: "In Progress"
      key: "doing"
    - name: "Review"
      key: "review"
    - name: "Done"
      key: "done"
  context:
    scope: ""
    release_target: ""
    owners: []
    notes: ""
next_id: 1
```

**Used when**: Creating new boards with `mochi-sticky board add`

## ADR Template

Located at `.sticky/templates/adr.md`:

```markdown
---
id: {{.ID}}
uid: "{{.UID}}"
title: "{{.Title}}"
status: "proposed"
date: {{.Date}}
tags: []
supersedes: []
superseded_by: []
links: []
---

## Context

Describe the problem and why a decision is needed.

## Decision

State the chosen approach and key implementation details.

## Consequences

**Positive**:
- Benefit 1
- Benefit 2

**Negative**:
- Trade-off 1
- Trade-off 2

## Related

Link to relevant ADRs, documentation, or discussions.
```

**Variables**:
- `{{.ID}}` — Sequential ADR number
- `{{.UID}}` — Generated UID with timestamp
- `{{.Title}}` — Title from command
- `{{.Date}}` — Current date (YYYY-MM-DD)

## Wiki Page Template

Located at `.sticky/templates/wiki/page.md`:

```markdown
---
title: {{.Title}}
slug: {{.Slug}}
section: {{.Section}}
order: {{.Order}}
tags: []
status: draft
---

# {{.Title}}

Page content goes here.

## Section 1

Content for section 1.

## Section 2

Content for section 2.
```

**Variables**:
- `{{.Title}}` — Page title
- `{{.Slug}}` — URL slug
- `{{.Section}}` — Wiki section
- `{{.Order}}` — Sort order in section

## PDF Template

Located at `.sticky/templates/wiki/wiki_pdf_template.tex`:

LaTeX template for PDF export via Pandoc. See [PDF Export](../export/pdf.md) for details.

**Key features**:
- Custom title page
- Table of contents
- Syntax highlighting
- Page headers/footers

## Template Configuration

Configure template locations in `.sticky/mochi-sticky.yaml`:

```yaml
templates:
  root: .mochi-sticky/templates
  task: .mochi-sticky/templates/task.md
  board: .mochi-sticky/templates/board.yaml
  adr: .mochi-sticky/templates/adr.md
  wiki: .mochi-sticky/templates/wiki
  wiki_pdf: .mochi-sticky/templates/wiki/wiki_pdf_template.tex
```

## Customizing Templates

### Modify Task Template

Edit `.sticky/templates/task.md`:

```markdown
---
id: "{{.ID}}"
uid: "{{.UID}}"
title: "{{.Title}}"
status: "{{.Status}}"
priority: 2
tags: []
created: {{.Created}}
dependencies: []
---

# {{.Title}}

## Problem Statement

What problem does this task solve?

## Proposed Solution

How will we solve it?

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2

## Testing Plan

How will we verify this works?

## Rollout Plan

How will we deploy this?
```

**Result**: All new tasks use this structure

### Customize Board Columns

Edit `.sticky/templates/board.yaml`:

```yaml
board:
  columns:
    - name: "Backlog"
      key: "backlog"
    - name: "Ready"
      key: "ready"
    - name: "Active"
      key: "active"
    - name: "Testing"
      key: "testing"
    - name: "Complete"
      key: "complete"
  context:
    scope: ""
    sprint_number: 0
    team: ""
next_id: 1
```

**Result**: New boards have custom workflow

### Customize ADR Template

Edit `.sticky/templates/adr.md`:

```markdown
---
id: {{.ID}}
uid: "{{.UID}}"
title: "{{.Title}}"
status: "proposed"
date: {{.Date}}
author: ""
reviewers: []
tags: []
supersedes: []
superseded_by: []
links: []
---

## Background

What led to this decision?

## Options Considered

### Option 1: [Name]
- Pros: ...
- Cons: ...

### Option 2: [Name]
- Pros: ...
- Cons: ...

## Decision

Which option was chosen and why?

## Implementation Plan

How will we execute this decision?

## Success Metrics

How will we measure success?
```

## Template Best Practices

- **Keep templates minimal**:
- Include only essential structure
- Let users add content as needed
- Avoid overwhelming new users

- **Use clear section headings**:
```markdown
## Acceptance Criteria
## Testing Plan
## Related Tasks
```

- **Include helpful comments**:
```markdown
## Notes

<!-- Add implementation notes, design decisions, or links -->
```

- **Provide examples inline**:
```yaml
tags: [backend, api]  # Example: topic-based tags
priority: 2           # 1=high, 2=normal, 3=low
```

- **Keep frontmatter flexible**:
```yaml
tags: []              # Empty array, easily extended
dependencies: []      # Optional relationships
```

- **Version control templates**:
```gitignore
# Commit templates
.sticky/templates/

# Ignore data
.sticky/boards/
.sticky/adrs/
.sticky/wiki/
```

## Troubleshooting

### Template Not Found

**Error**: "template not found"

**Solution**:
1. Check template exists:
   ```bash
   ls .sticky/templates/task.md
   ```

2. Verify configuration:
   ```bash
   cat .sticky/mochi-sticky.yaml
   ```

3. Re-initialize if missing:
   ```bash
   mochi-sticky init --force
   ```

### Invalid Template Variables

**Problem**: Variables not replaced (see literal `{{.ID}}`)

**Cause**: Template syntax error or wrong delimiters

**Solution**: Use Go template syntax:
```
{{.Variable}}  - Correct
${Variable}    ❌ Wrong
$Variable      ❌ Wrong
```

## Related

- [Tasks](tasks.md) — Task structure
- [Configuration](config.md) — Template paths
- [ADR](adr.md) — ADR structure
- [Wiki Schema](wiki-schema.md) — Wiki frontmatter
