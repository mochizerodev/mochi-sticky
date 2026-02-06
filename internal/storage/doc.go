// Package storage centralizes the logic for locating and loading the repository's
// persistent storage root (typically `.sticky`). It respects overrides via the
// environment, command line, and `.sticky/mochi-sticky.yaml`, and ensures the resolved
// path exists and is a directory before the rest of the application relies on it.
// The package also exposes helper functions for loading the config file and
// resolving arbitrary storage-related paths relative to the current working
// directory while honoring `allowMissing` semantics.
package storage
