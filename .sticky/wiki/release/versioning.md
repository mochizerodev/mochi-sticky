---
title: Versioning Plan
slug: release/versioning
section: Release
order: 10
tags:
    - release
    - versioning
status: published
---
# Versioning Plan

This project follows **Semantic Versioning (SemVer)**: `MAJOR.MINOR.PATCH`.

## Policy

- **0.y.z**: Early stage. Backward-incompatible changes are allowed while APIs and file formats stabilize.
- **1.0.0+**: Breaking changes increment **MAJOR**. New features increment **MINOR**. Fixes increment **PATCH**.
- The `.sticky/` data format is treated as part of the public API once 1.0.0 is reached.

## Release Tags

- Releases are created from Git tags: `vX.Y.Z`.
- The tag value becomes the user-facing version string.

## Build Stamping

Release binaries are built with version metadata injected at build time:

- Version: `mochi-sticky/internal/version.Version`
- Commit: `mochi-sticky/internal/version.Commit`
- Date: `mochi-sticky/internal/version.Date`

The release workflow sets these via `-ldflags` so `mochi-sticky version` and `mochi-sticky --version` report the tagged build.

## Local Builds

Local builds default to `Version=dev`, `Commit=none`, `Date=unknown`. You can override locally if needed:

```bash
go build -ldflags "-X mochi-sticky/internal/version.Version=dev-local -X mochi-sticky/internal/version.Commit=$(git rev-parse --short HEAD) -X mochi-sticky/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./...
```

## Changelog

The `CHANGELOG.md` should be updated with each release tag, and release notes should be generated using the wiki template at `release/release-notes-template`.
