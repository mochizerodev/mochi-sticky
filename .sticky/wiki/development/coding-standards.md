---
title: Coding Standards
slug: development/coding-standards
section: Development
order: 42
tags:
    - development
    - standards
    - go
status: published
---
# Coding Standards

The Mochi Way - coding standards for mochi-sticky contributors.

## Error Handling

### No Panics

**Never use `panic()` or `os.Exit()` inside `internal/`**

```go
// BAD
func LoadTask(id string) Task {
    if id == "" {
        panic("task ID cannot be empty")
    }
    // ...
}

// GOOD
func LoadTask(id string) (Task, error) {
    if id == "" {
        return Task{}, fmt.Errorf("task ID cannot be empty")
    }
    // ...
}
```

**Rationale**: Panics crash the entire application. Library code should always return errors and let the caller decide how to handle them.

### Error Wrapping

**Always wrap errors with context**

```go
// BAD
func parseTask(path string) (Task, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Task{}, err  // Lost context!
    }
    // ...
}

// GOOD
func parseTask(path string) (Task, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Task{}, fmt.Errorf("parse task: failed to read %s: %w", path, err)
    }
    // ...
}
```

**Format**: `"component: action failed: %w"`

**Examples**:
- `fmt.Errorf("board: failed to parse task %s: %w", id, err)`
- `fmt.Errorf("repository: list tasks: %w", err)`
- `fmt.Errorf("config: invalid column key %q: %w", key, err)`

### Sentinel Errors

**Define custom error types for common failures**

```go
// internal/board/errors.go
var (
    ErrTaskNotFound   = errors.New("task not found")
    ErrBoardNotFound  = errors.New("board not found")
    ErrInvalidStatus  = errors.New("invalid status")
    ErrCircularDep    = errors.New("circular dependency detected")
)
```

**Usage**:
```go
func GetTask(id string) (Task, error) {
    task, exists := tasks[id]
    if !exists {
        return Task{}, ErrTaskNotFound
    }
    return task, nil
}

// Caller can check
task, err := GetTask("T-000042")
if errors.Is(err, board.ErrTaskNotFound) {
    // Handle gracefully
}
```

## Concurrency & State

### Thread Safety

**Repository must be thread-safe using `sync.RWMutex`**

```go
type Repository struct {
    mu    sync.RWMutex
    tasks map[string]*Task
}

func (r *Repository) GetTask(id string) (Task, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    task, exists := r.tasks[id]
    if !exists {
        return Task{}, ErrTaskNotFound
    }
    return *task, nil
}

func (r *Repository) UpdateTask(task Task) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.tasks[task.ID] = &task
    return r.saveTask(task)
}
```

**Rules**:
- Use `RLock` for read operations
- Use `Lock` for write operations
- Always defer unlock immediately after acquiring lock

### TUI Concurrency

**Never update model directly from goroutine - use `tea.Cmd`**

```go
// BAD
func (m Model) Init() tea.Cmd {
    go func() {
        tasks := loadTasks()
        m.tasks = tasks  // Race condition!
    }()
    return nil
}

// GOOD
type tasksLoadedMsg struct {
    tasks []Task
}

func loadTasksCmd() tea.Cmd {
    return func() tea.Msg {
        tasks := loadTasks()
        return tasksLoadedMsg{tasks: tasks}
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tasksLoadedMsg:
        m.tasks = msg.tasks
    }
    return m, nil
}
```

## Code Style

### Minimalist Functions

**Keep functions small and focused on a single task**

```go
// BAD - Does too much
func ProcessTask(id string) error {
    task := loadTask(id)
    validateTask(task)
    updateStatus(task)
    saveTask(task)
    notifyUsers(task)
    logActivity(task)
    updateMetrics(task)
    return nil
}

// GOOD - Single responsibility
func LoadTask(id string) (Task, error) { /* ... */ }
func ValidateTask(task Task) error { /* ... */ }
func UpdateStatus(task *Task, status string) error { /* ... */ }
```

## Tooling

### Pre-commit

We use `pre-commit` to run `golangci-lint` before commits.

Install:
```bash
# macOS (Homebrew)
brew install pre-commit

# Ubuntu/Debian
sudo apt-get install pre-commit

# Python (any OS)
python -m pip install --user pre-commit
```

Enable the hook:
```bash
pre-commit install
```

Run hooks manually:
```bash
pre-commit run --all-files
```

**Target**: < 30 lines per function (excluding comments)

### Self-Documenting Code

**Use descriptive variable names**

```go
// BAD
func p(tp string) (T, error) {
    d, e := os.ReadFile(tp)
    if e != nil {
        return T{}, e
    }
    // ...
}

// GOOD
func parseTask(taskFilePath string) (Task, error) {
    data, err := os.ReadFile(taskFilePath)
    if err != nil {
        return Task{}, fmt.Errorf("parse task: %w", err)
    }
    // ...
}
```

**Guidelines**:
- Use full words: `taskID` not `tid`
- Be specific: `boardConfigPath` not `path`
- Avoid abbreviations except common ones: `err`, `ctx`, `id`

### Documentation

**Every exported function MUST have a comment starting with the function name**

```go
// BAD - Missing comment
func ListTasks(filter Filter) ([]Task, error) {
    // ...
}

// BAD - Doesn't start with function name
// Returns all tasks matching the filter
func ListTasks(filter Filter) ([]Task, error) {
    // ...
}

// GOOD
// ListTasks returns all tasks matching the provided filter criteria.
// Returns ErrBoardNotFound if the active board does not exist.
func ListTasks(filter Filter) ([]Task, error) {
    // ...
}
```

**Package documentation** (doc.go):
```go
// Package board provides task and board management functionality.
// It handles task creation, updates, archiving, and board operations
// with thread-safe repository access.
package board
```

### Interfaces

**Accept interfaces, return structs**

```go
// GOOD
type TaskRepository interface {
    GetTask(id string) (Task, error)
    SaveTask(task Task) error
}

func ProcessTask(repo TaskRepository, id string) (Task, error) {
    task, err := repo.GetTask(id)
    // ...
    return task, nil  // Return concrete type
}
```

**Rationale**: 
- Accepting interfaces makes testing easier (mock implementations)
- Returning structs is more efficient and clearer

## Testing

### Table-Driven Tests

**Use Go's table-driven pattern for all logic**

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
            input: `---
id: "T-000001"
title: "Test"
---`,
            want: Task{ID: "T-000001", Title: "Test"},
        },
        {
            name:    "missing ID",
            input:   `---\ntitle: "Test"\n---`,
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

### Mock File Systems

**Test Repository without hitting actual disk**

```go
type mockFS struct {
    files map[string][]byte
}

func (m *mockFS) ReadFile(path string) ([]byte, error) {
    data, exists := m.files[path]
    if !exists {
        return nil, os.ErrNotExist
    }
    return data, nil
}

func TestRepository_GetTask(t *testing.T) {
    fs := &mockFS{
        files: map[string][]byte{
            "tasks/T-000001.md": []byte(`---
id: "T-000001"
title: "Test Task"
---`),
        },
    }
    
    repo := NewRepository(fs)
    task, err := repo.GetTask("T-000001")
    // ...
}
```

### AAA Pattern

**Structure tests: Arrange, Act, Assert**

```go
func TestUpdateTaskStatus(t *testing.T) {
    // Arrange
    task := Task{
        ID:     "T-000001",
        Status: "todo",
    }
    
    // Act
    err := UpdateStatus(&task, "doing")
    
    // Assert
    if err != nil {
        t.Fatalf("UpdateStatus() error = %v", err)
    }
    if task.Status != "doing" {
        t.Errorf("Status = %v, want %v", task.Status, "doing")
    }
}
```

## Output & Logging

### Stdout is Forbidden

**NEVER use `fmt.Println` or `log.Println` for debugging**

```go
// ABSOLUTELY FORBIDDEN
func LoadTasks() []Task {
    fmt.Println("Loading tasks...")  // Breaks TUI!
    log.Println("Debug info")        // Breaks MCP!
    // ...
}
```

**Why?**
- **In TUI**: Corrupts the UI buffer (Bubble Tea controls stdout)
- **In MCP**: Breaks JSON-RPC pipe (agents expect only JSON on stdout)

### Logging Strategy

**TUI debugging - use file logging**

```go
// In TUI initialization
if os.Getenv("DEBUG") != "" {
    f, err := tea.LogToFile("debug.log", "mochi-sticky")
    if err != nil {
        // Can't log, continue anyway
    }
    defer f.Close()
}

// In TUI code
log.Println("Selected task:", taskID)  // Goes to debug.log
```

**Core logic - accept Logger interface**

```go
type Logger interface {
    Printf(format string, args ...interface{})
}

type Repository struct {
    logger Logger
    // ...
}

func (r *Repository) LoadTask(id string) (Task, error) {
    if r.logger != nil {
        r.logger.Printf("Loading task %s", id)
    }
    // ...
}

// Usage
repo := NewRepository(logger)  // Can be file, stderr, or discard
```

### Stderr Usage

**Reserved for fatal application errors before TUI/MCP starts**

```go
// OK - Fatal startup error
func main() {
    if err := validateConfig(); err != nil {
        fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
        os.Exit(1)
    }
    
    // Start TUI - no more stderr output
    runTUI()
}
```

## File System Hygiene

### Cross-Platform Paths

**Never hardcode path separators**

```go
// BAD
taskPath := boardDir + "/tasks/" + id + ".md"

// GOOD
taskPath := filepath.Join(boardDir, "tasks", id+".md")
```

**Always use `filepath` package**:
- `filepath.Join()` - Build paths
- `filepath.Dir()` - Get directory
- `filepath.Base()` - Get filename
- `filepath.Ext()` - Get extension

### Windows Compatibility

**Handle Windows newlines (`\r\n`)**

```go
// GOOD - Normalize line endings
func parseMarkdown(content string) (string, error) {
    // Replace Windows CRLF with LF
    content = strings.ReplaceAll(content, "\r\n", "\n")
    
    // Now parse with \n only
    lines := strings.Split(content, "\n")
    // ...
}
```

### Git Ignore

**Auto-create `.gitignore` for debug files**

```go
func ensureGitIgnore(storageRoot string) error {
    gitignorePath := filepath.Join(storageRoot, ".gitignore")
    
    // Check if exists
    if _, err := os.Stat(gitignorePath); err == nil {
        return nil  // Already exists
    }
    
    content := `# mochi-sticky debug files
debug.log
*.log
`
    
    return os.WriteFile(gitignorePath, []byte(content), 0644)
}
```

## Code Review Checklist

Before submitting a PR, verify:

- [ ] No `panic()` or `os.Exit()` in `internal/`
- [ ] All errors wrapped with context
- [ ] Exported functions have doc comments
- [ ] Tests added for new functionality
- [ ] Table-driven tests where appropriate
- [ ] No `fmt.Println` or `log.Println` to stdout
- [ ] Cross-platform paths using `filepath`
- [ ] Thread-safe if accessing shared state
- [ ] TUI updates via `tea.Cmd` only
- [ ] Functions < 30 lines
- [ ] Descriptive variable names
- [ ] All tests pass: `go test ./...`

## Tools & Automation

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### Formatting

```bash
# Format all code
go fmt ./...

# Or use gofmt
gofmt -s -w .
```

### Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...

# Specific package
go test ./internal/board/...
```

## Examples

### Good Error Handling

```go
func (r *Repository) GetTask(id string) (Task, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    task, exists := r.tasks[id]
    if !exists {
        return Task{}, fmt.Errorf("get task: %w", ErrTaskNotFound)
    }
    
    return *task, nil
}
```

### Good Test Structure

```go
func TestRepository_UpdateTask(t *testing.T) {
    tests := []struct {
        name    string
        task    Task
        wantErr bool
    }{
        {
            name: "valid task",
            task: Task{ID: "T-000001", Title: "Test"},
        },
        {
            name:    "empty ID",
            task:    Task{Title: "Test"},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            repo := NewTestRepository(t)
            
            // Act
            err := repo.UpdateTask(tt.task)
            
            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Good Documentation

```go
// Package board provides task and board management with file-based storage.
//
// The main types are Task and Board, representing individual work items
// and their organizational containers. The Repository type handles
// persistent storage and retrieval with thread-safe operations.
package board

// Task represents a single work item with metadata and content.
type Task struct {
    ID       string    // Unique identifier (T-NNNNNN)
    Title    string    // Short description
    Status   string    // Current state (must match board column)
    Priority int       // Urgency level (1=high, 2=normal, 3=low)
    Created  time.Time // Creation timestamp
}

// NewTask creates a task with the provided title and default values.
// Returns an error if title is empty.
func NewTask(title string) (Task, error) {
    if title == "" {
        return Task{}, errors.New("title cannot be empty")
    }
    
    return Task{
        ID:       generateID(),
        Title:    title,
        Status:   "todo",
        Priority: 2,
        Created:  time.Now(),
    }, nil
}
```

## Related

- [Contributing Guide](contributing.md) — How to contribute
- [Architecture](architecture.md) — System design
- [Testing Guide](testing.md) — Testing practices
- [MASTER_SPEC.md](../../../MASTER_SPEC.md) — Complete specification
