package version

import (
	"fmt"
	"runtime"
)

// Version metadata is injected at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a detailed version string for human-readable output.
func String() string {
	return fmt.Sprintf(
		"mochi-sticky %s (commit %s, built %s, %s)",
		Version,
		Commit,
		Date,
		runtime.Version(),
	)
}

// Short returns the semantic version string only.
func Short() string {
	return Version
}
