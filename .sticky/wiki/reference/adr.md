---
title: Architecture Decision Records (ADR)
slug: reference/adr
section: Reference
order: 25
tags:
    - adr
    - architecture
status: published
---
# Architecture Decision Records (ADR)

ADRs document important architectural decisions made during development.

## File Location

`.sticky/adrs/*.md`

## ADR Structure

### Complete Example

```markdown
---
id: 4
uid: "adr-0004-20260204123045"
title: "Use JWT for API Authentication"
status: "accepted"
date: 2026-02-04
tags:
  - security
  - api
  - auth
supersedes:
  - 2
superseded_by: []
links:
  - title: "JWT RFC"
    url: "https://tools.ietf.org/html/rfc7519"
---

## Context

We need a stateless authentication mechanism for our microservices architecture that scales horizontally without session storage.

## Decision

Adopt JWT (JSON Web Tokens) for API authentication with:
- Short-lived access tokens (15 minutes)
- Long-lived refresh tokens (7 days)
- RS256 signing algorithm

## Consequences

**Positive**:
- Stateless authentication
- Easy horizontal scaling
- Standard format

**Negative**:
- Cannot revoke tokens before expiry
- Token size larger than session IDs
- Requires robust key management

## Status

Accepted on 2026-02-04

## Related

Supersedes: ADR-0002 (Session-based Authentication)
```

## Frontmatter Fields

### id (required)

Sequential 4-digit ADR number.

**Format**: Integer (e.g., `4`, `42`)

**Display**: Zero-padded in filenames (`ADR-0004.md`)

### uid (required)

Unique identifier with timestamp.

**Format**: `adr-NNNN-YYYYMMDDHHmmss`

**Example**: `adr-0004-20260204123045`

### title (required)

Decision summary starting with a verb.

**Best practices**:
- Start with action verb (Use, Adopt, Implement, Replace)
- Be specific and clear
- Keep under 80 characters

**Examples**:
```yaml
title: "Use PostgreSQL for primary database"
title: "Adopt microservices architecture"
title: "Replace REST with GraphQL API"
```

### status (required)

Current ADR state.

**Valid values**:
- `proposed` — Under consideration
- `accepted` — Approved and being implemented
- `deprecated` — No longer recommended
- `superseded` — Replaced by newer ADR

**Workflow**:
```
proposed → accepted → deprecated/superseded
```

### date (required)

Decision date.

**Format**: `YYYY-MM-DD`

**Example**: `2026-02-04`

### tags (optional)

Classification labels.

**Format**: Array of strings

**Common tags**:
```yaml
tags: [architecture, database, backend]
tags: [frontend, performance]
tags: [security, auth, api]
```

### supersedes (optional)

ADR IDs this decision replaces.

**Format**: Array of integers

**Example**:
```yaml
supersedes:
  - 2
  - 3
```

**Effect**: Referenced ADRs automatically marked as `superseded`

### superseded_by (optional)

ADR IDs that replace this decision (auto-populated).

**Format**: Array of integers

**Example**:
```yaml
superseded_by:
  - 7
```

### links (optional)

Related resources and documentation.

**Format**: Array of objects with `title` and `url`

**Example**:
```yaml
links:
  - title: "JWT RFC 7519"
    url: "https://tools.ietf.org/html/rfc7519"
  - title: "OAuth 2.0 Spec"
    url: "https://oauth.net/2/"
```

## ADR Content Structure

Follow this template:

### Context

Describe the problem and constraints:
- What decision needs to be made?
- Why is it needed now?
- What constraints exist?

### Decision

State the chosen solution clearly:
- What approach was selected?
- Key implementation details
- Specific technologies/patterns

### Consequences

Document trade-offs:

**Positive**:
- Benefits gained
- Problems solved
- Improvements enabled

**Negative**:
- Limitations introduced
- Technical debt created
- Risks accepted

## ADR Lifecycle

### 1. Creation

```bash
mochi-sticky adr create "Use PostgreSQL for primary database"
```

**Generated**:
- Sequential ID
- UID with timestamp
- Status: `proposed`
- Template structure
- Opens in editor

### 2. Editing

```bash
# Edit ADR content
mochi-sticky adr edit 4

# Move ADR (change ID)
mochi-sticky adr move 4 10
```

### 3. Status Changes

```bash
# List valid statuses
mochi-sticky adr statuses

# Manually change status by editing file
# or via supersession
```

###4. Supersession

```bash
# Create new ADR that supersedes ADR-0004
mochi-sticky adr create "Migrate to MongoDB" --supersedes 4
```

**Effect**:
- New ADR has `supersedes: [4]`
- ADR-0004 automatically gets `superseded_by: [<new-id>]`
- ADR-0004 status changes to `superseded`

## ADR Commands

### Create

```bash
# Basic creation
mochi-sticky adr create "Title"

# With tags
mochi-sticky adr create "Title" --tags database,backend

# Superseding another ADR
mochi-sticky adr create "New approach" --supersedes 4
```

### List

```bash
# All ADRs
mochi-sticky adr list

# Filter by status
mochi-sticky adr list --status accepted

# Filter by tag
mochi-sticky adr list --tag security

# Sort by date
mochi-sticky adr list --sort date --desc
```

### View

```bash
# View specific ADR
mochi-sticky adr view 4

# View with metadata
mochi-sticky adr view 4 --metadata
```

### Edit

```bash
# Edit in default editor
mochi-sticky adr edit 4

# Use specific editor
mochi-sticky adr edit 4 --editor vim
```

### Move

```bash
# Renumber ADR
mochi-sticky adr move 4 10

# Renumbers ADR-0004 to ADR-0010
```

### Lint

```bash
# Check all ADRs for issues
mochi-sticky adr lint

# Issues detected:
# - Missing required fields
# - Invalid status values
# - Broken supersedes references
# - Invalid date formats
```

## ADR Best Practices

- **Document before implementing**:
- Create ADR when considering options
- Status starts as `proposed`
- Accept after team review

- **Be specific about decisions**:
```markdown
# Good
Use PostgreSQL 14+ with JSONB for flexible schema parts

# Too vague
Use a database
```

- **Document trade-offs honestly**:
- Include negative consequences
- Acknowledge technical debt
- Note mitigation strategies

- **Link to relevant resources**:
```yaml
links:
  - title: "Benchmark results"
    url: "https://wiki.internal/benchmarks"
  - title: "Team discussion"
    url: "https://github.com/org/repo/issues/42"
```

- **Keep ADRs immutable**:
- Don't edit after acceptance
- Create new ADR to change course
- Use supersession chain

- **Tag consistently**:
```yaml
tags: [architecture, database, backend]  # Clear categories
```

- **Use supersession for evolution**:
```bash
# Original decision
ADR-0004: Use MySQL

# Later evolution
ADR-0012: Migrate to PostgreSQL
  supersedes: [4]
```

## Finding ADRs

### By Status

```bash
mochi-sticky adr list --status proposed   # Under review
mochi-sticky adr list --status accepted   # Active decisions
mochi-sticky adr list --status deprecated # Old approaches
```

### By Tag

```bash
mochi-sticky adr list --tag database
mochi-sticky adr list --tag security --tag api
```

### By Date

```bash
mochi-sticky adr list --sort date --desc  # Newest first
mochi-sticky adr list --from 2026-01-01   # This year
```

### Supersession Chain

```bash
# View ADR-0004
mochi-sticky adr view 4

# Check superseded_by field for newer decisions
```

## ADR Configuration

Located at `.sticky/adrs/config.yaml`:

```yaml
config_version: 1
next_id: 13
columns:
  - key: proposed
    title: Proposed
  - key: accepted
    title: Accepted
  - key: deprecated
    title: Deprecated
  - key: superseded
    title: Superseded
```

See [Schema Reference](schema.md) for the full ADR config schema.

### Custom Statuses

Add project-specific states:

```yaml
config_version: 1
columns:
  - key: proposed
    title: Proposed
  - key: review       # Added: Team review phase
    title: Review
  - key: accepted
    title: Accepted
  - key: implementing # Added: In progress
    title: Implementing
  - key: complete     # Added: Fully implemented
    title: Complete
  - key: deprecated
    title: Deprecated
  - key: superseded
    title: Superseded
```

## Troubleshooting

### Invalid Status

**Error**: "invalid status: xyz"

**Solution**:
```bash
mochi-sticky adr statuses  # List valid values
# Edit ADR to use valid status
```

### Broken Supersession Chain

**Problem**: ADR references non-existent ADR

**Solution**:
```bash
# Lint finds broken references
mochi-sticky adr lint

# Fix by editing frontmatter
mochi-sticky adr edit 4
# Remove invalid ID from supersedes array
```

### Missing Required Fields

**Error**: Lint reports missing fields

**Solution**: Edit ADR and add:
```yaml
id: 4
title: "Decision title"
status: "proposed"
date: 2026-02-04
```

## Related

- [Configuration](config.md) — ADR configuration
- [Wiki Schema](wiki-schema.md) — Similar frontmatter structure
- [Architecture](../development/architecture.md) — System design decisions
- [Contributing](../development/contributing.md) — How to propose changes
