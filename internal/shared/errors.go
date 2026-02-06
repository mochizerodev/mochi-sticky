package shared

import "errors"

// ErrInvalidPath indicates a path is outside the allowed sandbox.
var ErrInvalidPath = errors.New("invalid path")
