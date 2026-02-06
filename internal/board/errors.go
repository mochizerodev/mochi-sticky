package board

import "errors"

var (
	// ErrInvalidFrontmatter indicates the frontmatter section is missing or malformed.
	ErrInvalidFrontmatter = errors.New("invalid frontmatter")
	// ErrInvalidYAML indicates the frontmatter YAML could not be unmarshaled.
	ErrInvalidYAML = errors.New("invalid yaml")
	// ErrTaskNotFound indicates a task with the given ID does not exist.
	ErrTaskNotFound = errors.New("task not found")
	// ErrStoreNotInitialized indicates the .sticky/tasks directory is missing.
	ErrStoreNotInitialized = errors.New("store not initialized")
	// ErrInvalidID indicates a task ID is empty or contains invalid characters.
	ErrInvalidID = errors.New("invalid task id")
	// ErrBoardNotFound indicates a board with the given ID does not exist.
	ErrBoardNotFound = errors.New("board not found")
	// ErrInvalidBoardID indicates a board ID is empty or contains invalid characters.
	ErrInvalidBoardID = errors.New("invalid board id")
	// ErrBoardDeleteForbidden indicates a board cannot be deleted in its current state.
	ErrBoardDeleteForbidden = errors.New("board delete forbidden")
	// ErrInvalidTitle indicates a task title is empty or invalid.
	ErrInvalidTitle = errors.New("invalid title")
	// ErrInvalidPriority indicates a task priority is invalid.
	ErrInvalidPriority = errors.New("invalid priority")
	// ErrInvalidDependency indicates dependency list is invalid (cycle or bad id).
	ErrInvalidDependency = errors.New("invalid dependency")
)
