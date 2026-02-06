---
title: Testing
slug: development/testing
section: Development
order: 42
tags:
    - testing
status: published
---
# Testing

mochi-sticky maintains high test coverage to ensure reliability and prevent regressions. This guide covers testing strategies, running tests, and writing new tests.

## Quick Start

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/board/...

# Run specific test
go test -run TestParserParse ./internal/board/

# Skip cache (force rerun)
go test -count=1 ./...
```

## Test Organization

### Package Structure

```
internal/
├── board/
│   ├── parser.go
│   ├── parser_test.go          # Unit tests
│   ├── repository.go
│   ├── repository_test.go
│   └── test_helpers_test.go    # Shared test utilities
├── integration/
│   └── integration_test.go     # End-to-end tests
└── mcp/
    ├── server.go
    └── server_test.go
```

### Test Types

**Unit Tests** (`*_test.go`)
- Test individual functions and methods
- Use table-driven tests
- Mock external dependencies
- Fast execution (<1s per package)

**Integration Tests** (`internal/integration/`)
- Test component interactions
- Use real file system operations
- Validate end-to-end workflows
- Slower but comprehensive

**Property Tests** (planned)
- Round-trip serialization validation
- Invariant checking
- Fuzzing for parsers

## Writing Tests

### The AAA Pattern

All tests should follow the **Arrange-Act-Assert** pattern for clarity:

```go
func TestUpdateTaskStatus(t *testing.T) {
    // Arrange - Set up test data and dependencies
    repo := NewRepository(t.TempDir())
    task := &Task{
        ID:     "T-001",
        Title:  "Test",
        Status: "todo",
    }
    
    // Act - Execute the operation being tested
    updated, err := repo.UpdateTaskStatus(task.ID, "doing")
    
    // Assert - Verify the results
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if updated.Status != "doing" {
        t.Errorf("status = %q, want %q", updated.Status, "doing")
    }
}
```

**Benefits:**
- **Readability** — Clear test structure, easy to understand intent
- **Maintainability** — Changes are easier to locate
- **Consistency** — All tests follow the same pattern

Use comments to mark sections in complex tests:
```go
func TestComplexOperation(t *testing.T) {
    // Arrange
    repo := setupRepository(t)
    task := createTestTask(t, repo)
    
    // Act
    result := performOperation(task)
    
    // Assert
    assertExpectedResult(t, result)
}// Arrange
            parser := &Parser{}
            
            // Act
            got, err := parser.Parse(tt.input)
            
            // Assert
### Table-Driven Tests

The preferred pattern for mochi-sticky (AAA applied per iteration):

```go
func TestParseTask(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Task
        wantErr error
    }{
        {
            name: "valid task",
            input: `---
id: "T-001"
title: "Test"
status: "todo"
priority: 1
tags: [backend]
created: 2026-01-29
---
# Description
Task body`,
            want: &Task{
                ID:       "T-001",
                Title:    "Test",
                Status:   "todo",
                Priority: 1,
                Tags:     []string{"backend"},
                Content:  "# Description\nTask body",
            },
            wantErr: nil,
        },
        {
            name:    "missing frontmatter",
            input:   "No frontmatter",
            want:    nil,
            wantErr: ErrInvalidFrontmatter,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := &Parser{}
            got, err := parser.Parse(tt.input)
            
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("Parse() error = %v, want %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Helpers

Use helper functions for setup and assertions:

```go
// test_helpers_test.go
func requireNoError(t *testing.T, err error, msg string) {
    t.Helper()
    if err != nil {
        t.Fatalf("%s: unexpected error: %v", msg, err)
    }
}

func setupTestDir(t *testing.T) string {
    t.Helper()
    dir := t.TempDir() // Automatic cleanup
    return dir
}

func writeTestFile(t *testing.T, path, content string) {
    t.Helper()
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("failed to write test file: %v", err)
    }
}
```

### Testing File Operations

Always Arrange
    tmpDir := t.TempDir() // Cleaned up automatically
    repo := NewRepository(tmpDir)
    task := &Task{
        Title:    "Test Task",
        Status:   "todo",
        Priority: 1,
    }
    
    // Act
    created, err := repo.CreateTask(task)
    
    // Assert
    if err != nil {
        t.Fatalf("CreateTask failed: %v", err)
    }
    if created.ID == "" {
        t.Error("expected ID to be generated")
    }
    
    // Verify file exists on disk
        t.Error("expected ID to be generated")
    }
    
    // Verify file exists
    path := filepath.Join(tmpDir, "tasks", created.ID+".md")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Errorf("task file not created at %s", path)
    }
}
```

### Testing Errors

Test both success and failure cases:

```go
func TestValidateTask(t *testing.T) {
    tests := []struct {
        name    string
        task    *Task
        wantErr error
    }{
        {
            name: "valid task",
            task: &Task{
                Title:    "Valid",
                Status:   "todo",
                Priority: 1,
            },
            wantErr: nil,
        },
        {
            name: "empty title",
            task: &Task{
                Title:  "",
                Status: "todo",
            },
            wantErr: ErrEmptyTitle,
        },
        {
            name: "invalid priority",
            task: &Task{
                Title:    "Test",
                Priority: 5, // Invalid: must be 1-3
            },
            wantErr: ErrInvalidPriority,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateTask(tt.task)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got error %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

## Integration Tests

Test complete workflows:

```goArrange
        storageRoot := t.TempDir()
        repo := board.NewRepository(storageRoot)
        
        b, err := repo.CreateBoard("Test Board")
        requireNoError(t, err, "create board")
        
        task, err := repo.CreateTask(&board.Task{
            Title:    "Implement feature",
            Status:   "todo",
            Priority: 1,
            Tags:     []string{"backend"},
        })
        requireNoError(t, err, "create task")
        
        // Act - Execute the workflow
        moved, err := repo.UpdateTaskStatus(task.ID, "doing")
        requireNoError(t, err, "move task")
        
        err = repo.ArchiveTask(task.ID)
        requireNoError(t, err, "archive task")
        
        // Assert - Verify final state
        if moved.Status != "doing" {
            t.Errorf("status = %q, want %q", moved.Status, "doing")
        }
        Task(task.ID)
        requireNoError(t, err, "archive task")
        
        // Verify archived
        archived, err := repo.ListArchivedTasks()
        requireNoError(t, err, "list archived")
        
        if len(archived) != 1 {
            t.Errorf("got %d archived tasks, want 1", len(archived))
        }
    })
}
```

## Test Coverage

### Running Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Coverage by package
go test -cover ./...
```

### Coverage Goals

- **Domain logic** (`internal/board/`): 80%+ coverage
- **Parsers**: 90%+ coverage (critical for data integrity)
- **MCP server**: 70%+ coverage
- **CLI commands**: Basic smoke tests
- **TUI**: Manual testing (Bubble Tea limitations)

### Excluded from Coverage

- `cmd/` — Thin wrappers, tested via integration
- `internal/tui/` — Interactive UI, manual testing
- Error messages and formatting code
- Debug/development-only code

## Testing Best Practices

### DO

- **Use table-driven tests** for multiple scenarios
```go
tests := []struct{ name, input, want string }{...}
```
Follow the AAA pattern** for test structure
```go
// Arrange
setup := createTestData()
// Act
result := performOperation(setup)
// Assert
assertResult(t, result)
```

- **
- **Use t.Helper()** in helper functions
```go
func assertNoError(t *testing.T, err error) {
    t.Helper() // Better stack traces
    if err != nil {
        t.Fatal(err)
    }
}
```

- **Use t.TempDir()** for file system tests
```go
dir := t.TempDir() // Auto-cleanup
```

- **Test error cases** explicitly
```go
if err == nil {
    t.Error("expected error, got nil")
}
```

- **Use subtests** for organization
```go
t.Run("ValidInput", func(t *testing.T) { ... })
```

### DON'T

❌ **Don't use `t.Fatal` in goroutines** (causes panic)
```go
// Bad
go func() {
    t.Fatal("error") // PANIC!
}()

// Good
go func() {
    if err != nil {
        errChan <- err
    }
}()
```

❌ **Don't test implementation details** (test behavior)
```go
// Bad: testing private fields
if task.internalState == "foo" { ... }

// Good: testing observable behavior  
if task.Status != "todo" { ... }
```

❌ **Don't share state between tests**
```go
// Bad: global variable
var sharedRepo *Repository

// Good: create new instance per test
repo := NewRepository(t.TempDir())
```

❌ **Don't use hard-coded paths**
```go
// Bad
repo := NewRepository("/tmp/test")

// Good
repo := NewRepository(t.TempDir())
```

## Mocking

### Interface-Based Mocking

```go
// repository.go
type FileReader interface {
    ReadFile(path string) ([]byte, error)
}

// repository_test.go
type mockFileReader struct {
    content string
    err     error
}

func (m *mockFileReader) ReadFile(path string) ([]byte, error) {
    return []byte(m.content), m.err
}

func TestWithMock(t *testing.T) {
    mock := &mockFileReader{
        content: "test content",
        err:     nil,
    }
    
    // Use mock in test
    result := ProcessFile(mock)
    // assertions...
}
```

## Race Detection

Always test with race detector:

```bash
# Run with race detection
go test -race ./...

# CI should always use this
go test -race -count=1 ./...
```

Common race conditions to watch for:
- Concurrent map access
- Shared repository state
- TUI model updates from goroutines

## Benchmarking

Performance-critical code should have benchmarks:

```go
func BenchmarkParseTask(b *testing.B) {
    parser := &Parser{}
    input := `---
id: "T-001"
title: "Benchmark Task"
status: "todo"
---
Body content`
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = parser.Parse(input)
    }
}

// Run benchmarks
// go test -bench=. ./internal/board/
// go test -bench=BenchmarkParseTask -benchmem
```

## Continuous Integration

### Pre-commit Checks

```bash
#!/bin/bash
# .git/hooks/pre-commit

set -e

echo "Running tests..."
go test ./...

echo "Running race detector..."
go test -race ./...

echo "Checking formatting..."
gofmt -l . | grep -v '^$' && exit 1

echo "All checks passed!"
```

### CI Pipeline

```yaml
# .github/workflows/test.yml (example)
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -race -count=1 -timeout 5m ./...
      - run: go test -cover ./...
```

## Debugging Tests

### Verbose Output

```bash
# See all test output
go test -v ./...

# See only failures
go test ./...
```

### Specific Tests

```bash
# Run one test
go test -run TestParseTask ./internal/board/

# Run tests matching pattern
go test -run Parse ./...
```

### Test Timeouts

```bash
# Prevent hanging tests
go test -timeout 30s ./...

# Short mode (skip slow tests)
go test -short ./...
```

### Debugging Tips

```go
// Use t.Log for debugging (only shown on failure)
t.Logf("intermediate value: %v", x)

// Use t.Error vs t.Fatal appropriately
t.Error("continues test")   // Non-fatal
t.Fatal("stops test")        // Fatal

// Print structured data
t.Logf("task: %+v", task)    // Print with field names
```

## Testing Checklist

When adding new features:

- [ ] Unit tests for new functions
- [ ] Table-driven tests for multiple scenarios
- [ ] Error case coverage
- [ ] Edge cases (empty input, nil, boundaries)
- [ ] Integration test for workflow
- [ ] Race detector passes
- [ ] Coverage maintained or improved
- [ ] Benchmarks for performance-critical code

## C// Arrange
    original := &Task{
        ID:       "T-001",
        Title:    "Test Task",
        Status:   "todo",
        Priority: 1,
    }
    
    // Act - Serialize and deserialize
    data, err := Render(original)
    requireNoError(t, err, "render")
    
    parsed, err := Parse(data)
    requireNoError(t, err, "parse")
    
    // Assert - Should match originalrror(t, err, "render")
    
    // Deserialize
    parsed, err := Parse(data)
    requireNoError(t, err, "parse")
    // Arrange
    data := createTestData()
    goldenPath := filepath.Join("testdata", "export.golden")
    
    // Act
    result := Export(data)
    
    // Assert - Compare with golden file    t.Errorf("round-trip failed: got %v, want %v", parsed, original)
    }
}
```

### Golden Files

```go
func TestExport(t *testing.T) {
    result := Export(data)
    
    goldenPath := filepath.Join("testdata", "export.golden")
    
    if *update {
        // Update golden file
        os.WriteFile(goldenPath, result, 0644)
    }
    
    golden, _ := os.ReadFile(goldenPath)
    if !bytes.Equal(result, golden) {
        t.Error("output doesn't match golden file")
    }
}
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Test Fixtures](https://pkg.go.dev/testing#hdr-Main)
- [Architecture Guide](architecture.md) — Understanding components to test
- [Contributing Guide](contributing.md) — Development workflow
