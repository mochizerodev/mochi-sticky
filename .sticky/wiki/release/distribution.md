---
title: Distribution
slug: release/distribution
section: Release
order: 0
tags: []
status: published
---
# Distribution Guide

This document describes release packaging, install artifacts, and CI validation for Linux, macOS, and Windows.

## Artifact naming

Release artifacts are produced by platform and architecture:

- Raw binaries:
  - `mochi-sticky-linux-amd64`
  - `mochi-sticky-linux-arm64`
  - `mochi-sticky-darwin-amd64`
  - `mochi-sticky-darwin-arm64`
  - `mochi-sticky-windows-amd64.exe`
- Install archives:
  - `mochi-sticky-linux-amd64.tar.gz`
  - `mochi-sticky-linux-arm64.tar.gz`
  - `mochi-sticky-darwin-amd64.tar.gz`
  - `mochi-sticky-darwin-arm64.tar.gz`
  - `mochi-sticky-windows-amd64.zip`
- Checksums:
  - `<artifact>.sha256` for each binary/archive
- Metadata:
  - `mochi-sticky-<goos>-<goarch>.metadata.json`

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
- Package installer archives (`.tar.gz` on Linux/macOS, `.zip` on Windows)
- Generate per-artifact checksums (`.sha256`)
- Verify install + version + basic command on Linux/macOS/Windows runners
- Publish release assets and release notes with installer snippets

## 4. CI pipeline behavior

- GitHub Actions builds artifacts for Linux/macOS/Windows and uploads them to the GitHub Release.
- Version metadata is stamped at build time via `-ldflags` (version, commit, build date).
- The workflow uploads per-platform install bundles as intermediate artifacts.
- A smoke test job downloads install bundles and validates:
  - checksum verification
  - installer execution
  - `mochi-sticky --version`
  - `mochi-sticky init` + `mochi-sticky board list`
- The release workflow emits structured telemetry JSON for build/smoke stages per platform.
- Telemetry files are written under `.sticky/release/telemetry/runs/` and uploaded as CI artifacts (and with release assets for build jobs).
- Release notes are generated with install snippets for shell/PowerShell installers and direct artifact names.

## 5. Release telemetry interpretation

- Telemetry fields include run identity, workflow/job metadata, git ref/commit, platform status, per-stage durations, and artifact byte sizes.
- Use `mochi-sticky release telemetry` to summarize recent runs (success rate + slowest stage).
- Use `mochi-sticky release telemetry --json` for machine-readable dashboards/automation.
- If telemetry was downloaded from CI/release artifacts, import it locally with:
  - `mochi-sticky release telemetry import path/to/telemetry.json`

## 6. Installer operations by platform

Linux:
- Install/upgrade/downgrade: rerun `scripts/install.sh` with optional `--version <tag>`.
- Verify: `mochi-sticky --version`, then `mochi-sticky init` + `mochi-sticky board list`.
- Uninstall: remove `mochi-sticky` from `/usr/local/bin` or `~/.local/bin`.

macOS:
- Install/upgrade/downgrade: shell installer (`scripts/install.sh`) and optional Homebrew flow when formula/tap is available.
- Verify: `mochi-sticky --version`, `mochi-sticky init`, basic command run.
- Uninstall: remove binary path (or `brew uninstall mochi-sticky` if installed via Homebrew).

Windows:
- Install/upgrade/downgrade: `scripts/install.ps1` using release `.zip` + `.sha256`.
- Verify: `mochi-sticky --version` and `mochi-sticky board list` in a fresh initialized directory.
- Uninstall: remove `%LOCALAPPDATA%\Programs\mochi-sticky\bin\mochi-sticky.exe`.

## 7. Failure handling reference

- PATH issues: ensure install directory is in PATH and restart shell/PowerShell session.
- Permission errors: install to user-local directory (`~/.local/bin` or `%LOCALAPPDATA%\Programs\...`) or rerun with elevated privileges.
- Checksum mismatches: ensure artifact and checksum come from the same release tag; retry download; avoid stale mirrored caches.

## 8. Signing

Binary signing is not implemented yet. If/when added, platform-specific signing steps should be inserted into the release workflow after the build step.

## 9. Optional: GoReleaser

If you adopt GoReleaser later, it can automate multi-platform builds, checksums, and GitHub Releases.
