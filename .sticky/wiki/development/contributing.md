---
title: Contributing
slug: development/contributing
section: Development
order: 41
tags:
    - contributing
status: published
---
# Contributing

Thank you for considering contributing to mochi-sticky! This guide will help you get started.

## Quick Start

### Prerequisites

- **Go 1.21+** — [Download](https://go.dev/dl/)
- **Git** — For version control
- **Make** (optional) — For build automation

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/mochi-sticky.git
cd mochi-sticky

# Add upstream remote
git remote add upstream https://github.com/ORIGINAL_OWNER/mochi-sticky.git
```

### Build

```bash
# Download dependencies
go mod download

# Build the binary
go build -o mochi-sticky

# Verify it works
./mochi-sticky --version
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/board/...
```

## Development Workflow

### 1. Create a Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Changes

- Write code following [Coding Standards](coding-standards.md)
- Add tests for new functionality
- Update documentation if needed

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run linter (if available)
golangci-lint run

# Format code
go fmt ./...

# Build to ensure it compiles
go build -o mochi-sticky
```

### 4. Commit

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add task dependency validation"
```

**Commit message format**:
- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation changes
- `test:` — Test additions/changes
- `refactor:` — Code refactoring
- `chore:` — Maintenance tasks

### 5. Push and Create PR

```bash
# Push to your fork
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub with:
- Clear title describing the change
- Description of what changed and why
- Link to related issues (if any)
- Screenshots (for UI changes)

## Coding Standards

**Please read and follow [Coding Standards](coding-standards.md) carefully.**

Key principles:

### Error Handling
- ❌ **No panics** in library code (`internal/`)
- ✅ **Wrap errors** with context: `fmt.Errorf("component: action: %w", err)`
- ✅ **Use sentinel errors** for common cases

### Code Style
- Keep functions **small and focused** (< 30 lines)
- Use **descriptive names**: `taskFilePath` not `tp`
- **Document all exports** with comments starting with function name
- **Accept interfaces, return structs**

### Output & Logging
- ❌ **NEVER** use `fmt.Println` or `log.Println` to stdout
  - Breaks TUI rendering
  - Breaks MCP JSON-RPC
- ✅ Use file logging for debugging
- ✅ Stderr only for fatal startup errors

### Testing
- Write **table-driven tests** for all logic
- Use **AAA pattern**: Arrange, Act, Assert
- Mock file system operations
- Aim for **>80% coverage** on new code

### Cross-Platform
- Use `filepath.Join()` for paths
- Handle Windows line endings (`\r\n`)
- Test on multiple platforms if possible

## What to Contribute

### Good First Issues

Look for issues labeled `good-first-issue`:
- Documentation improvements
- Small bug fixes
- Test coverage improvements
- Code examples

### Areas Needing Help

- **Tests** — Increase coverage
- **Documentation** — Tutorials, examples
- **Bug fixes** — Check GitHub issues
- **Performance** — Optimize hot paths
- **Features** — See roadmap in `roadmap/roadmap`

### Ideas Welcome

Have an idea? Open an issue first to discuss:
- Describe the problem it solves
- Explain your proposed solution
- Get feedback before implementing

## Testing Guidelines

### Unit Tests

Test individual functions in isolation:

```go
func TestParseTask(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Task
        wantErr bool
    }{
        {
            name: "valid task",
            input: "---\nid: T-000001\n---\n# Task",
            want: Task{ID: "T-000001"},
        },
        {
            name:    "invalid frontmatter",
            input:   "not yaml",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseTask(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseTask() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("parseTask() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Test component interactions:

```go
func TestRepository_Integration(t *testing.T) {
    // Setup test storage
    tmpDir := t.TempDir()
    repo := NewRepository(tmpDir)
    
    // Test complete workflow
    task, err := repo.CreateTask("Test task")
    if err != nil {
        t.Fatal(err)
    }
    
    // Update and verify
    task.Status = "doing"
    if err := repo.UpdateTask(task); err != nil {
        t.Fatal(err)
    }
    
    loaded, err := repo.GetTask(task.ID)
    if err != nil {
        t.Fatal(err)
    }
    
    if loaded.Status != "doing" {
        t.Errorf("Status = %v, want %v", loaded.Status, "doing")
    }
}
```

## Documentation

### Code Documentation

```go
// Package board provides task and board management.
package board

// Task represents a work item with metadata and content.
type Task struct {
    ID    string
    Title string
}

// NewTask creates a task with the provided title.
// Returns an error if title is empty.
func NewTask(title string) (Task, error) {
    // ...
}
```

### Wiki Documentation

Update wiki when adding features:

```bash
# Edit relevant wiki page
mochi-sticky wiki edit user-guide/cli

# Or create new page
mochi-sticky wiki create \
  --title "New Feature Guide" \
  --slug user-guide/new-feature \
  --section "User Guide"
```

## PR Review Process

### What We Look For

- ✅ Tests pass
- ✅ Code follows standards
- ✅ Documentation updated
- ✅ No stdout logging
- ✅ Errors properly wrapped
- ✅ Backward compatible (or breaking changes documented)

### Review Feedback

- Address all review comments
- Push new commits (don't force push during review)
- Mark conversations as resolved when fixed

### After Approval

- Maintainer will merge
- Your contribution will be in next release!

## Resources

### Documentation

- [Architecture](architecture.md) — System design overview
- [Coding Standards](coding-standards.md) — Detailed coding rules
- [Testing Guide](testing.md) — Testing best practices
- [MASTER_SPEC.md](../../../MASTER_SPEC.md) — Complete specification

### Community

- **GitHub Issues** — Bug reports and features
- **Pull Requests** — Code contributions
- **Discussions** — Questions and ideas

## Questions?

- Check existing documentation
- Search GitHub issues
- Open a new issue with `question` label
- Read [Architecture](architecture.md) for design decisions

## Thank You!

Every contribution helps make mochi-sticky better. Whether it's code, documentation, bug reports, or ideas — we appreciate your help! 🎉
