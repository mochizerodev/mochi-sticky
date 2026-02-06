package templates

import "embed"

//go:embed assets/**/*
var embeddedFS embed.FS
