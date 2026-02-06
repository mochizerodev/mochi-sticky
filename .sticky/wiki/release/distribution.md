---
title: Distribution
slug: release/distribution
section: Release
order: 0
tags: []
status: published
---
# Distribution Guide

This document describes how to build and release `mochi-sticky` for Linux, macOS, and Windows.

## Artifact naming

Release artifacts are named by platform and architecture:

- `mochi-sticky-linux-amd64`
- `mochi-sticky-linux-arm64`
- `mochi-sticky-darwin-amd64`
- `mochi-sticky-darwin-arm64`
- `mochi-sticky-windows-amd64.exe`

Windows `arm64` builds are currently excluded in CI.

## 1. Local builds

From the repo root:

Linux/macOS:
```bash
go build -o mochi-sticky
```

Windows (PowerShell):
```powershell
go build -o mochi-sticky.exe
```

## 2. Cross-compilation

Use `GOOS` and `GOARCH` for cross builds.

Linux (amd64/arm64):
```bash
GOOS=linux GOARCH=amd64 go build -o dist/mochi-sticky-linux-amd64
GOOS=linux GOARCH=arm64 go build -o dist/mochi-sticky-linux-arm64
```

macOS (amd64/arm64):
```bash
GOOS=darwin GOARCH=amd64 go build -o dist/mochi-sticky-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o dist/mochi-sticky-darwin-arm64
```

Windows (amd64/arm64):
```bash
GOOS=windows GOARCH=amd64 go build -o dist/mochi-sticky-windows-amd64.exe
GOOS=windows GOARCH=arm64 go build -o dist/mochi-sticky-windows-arm64.exe
```

## 3. Release checklist

- Run tests: `go test ./...`
- Build all targets into `dist/`
- Verify `mochi-sticky --version` output on each OS
- Create checksums (optional): `sha256sum dist/* > dist/sha256sums.txt`
- Tag the release in git, attach artifacts.

## 4. CI pipeline behavior

- GitHub Actions builds artifacts for Linux/macOS/Windows and uploads them to the GitHub Release.
- Version metadata is stamped at build time via `-ldflags` (version, commit, build date).
- A smoke test job runs after builds and executes `mochi-sticky --version` on each OS.

## 5. Signing

Binary signing is not implemented yet. If/when added, platform-specific signing steps should be inserted into the release workflow after the build step.

## 6. Optional: GoReleaser

If you adopt GoReleaser later, it can automate multi-platform builds, checksums, and GitHub Releases.
