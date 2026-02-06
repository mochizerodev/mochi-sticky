package cmd

import (
	"mochi-sticky/internal/cli"
	"mochi-sticky/internal/storage"
)

var storageRootFlag string

func init() {
	cli.SetStorageRootFlagRef(&storageRootFlag)
}

func resolveStorageRoot(workingDir string, allowMissing bool) (string, error) {
	return cli.ResolveStorageRoot(workingDir, allowMissing)
}

func loadStorageConfig(workingDir string) (storage.Config, error) {
	return cli.LoadStorageConfig(workingDir)
}
