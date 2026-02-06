---
title: Preparing a Release
slug: release/preparing-a-release
section: Release
order: 5
tags:
    - release
    - process
    - git
status: published
---
# Preparing a Release

This guide explains the recommended release process for `mochi-sticky`, including a branching flow that works well with a protected `main` branch.

If you haven’t yet, read:
- `release/versioning` (tagging + build stamping)
- `release/distribution` (artifacts + CI behavior)

## Overview

- Releases are published from a git tag: `vX.Y.Z`.
- GitHub Actions builds multi-platform artifacts and uploads them to the GitHub Release when the release is **published**.
- The release workflow generates release notes from the matching section in `CHANGELOG.md`.

## Recommended branching model

- `main`: protected, always releasable.
- Feature work: short-lived branches (e.g. `feat/...`, `fix/...`) merged via PR.
- Release prep: a dedicated branch named `prepare-release/vX.Y.Z`.

Why a release prep branch?
- It keeps release-only changes (changelog, docs, workflow tweaks) reviewable.
- It avoids mixing release housekeeping with feature work.

## Step-by-step release prep (protected `main`)

1. **Sync `main` locally**
   - `git checkout main && git pull`

2. **Create a release prep branch**
   - `git checkout -b prepare-release/vX.Y.Z`

3. **Update `CHANGELOG.md`**
   - Add a new section header that exactly matches the tag you will publish:
     - `## [vX.Y.Z]` (or `## vX.Y.Z`)
   - Move/copy the relevant items from `## [Unreleased]` into the new section.

4. **Optional: update docs/release messaging**
   Typical candidates:
   - `README.md` (platform notes, install notes, sponsor/support info)
   - `.sticky/wiki/release/*` docs if the process changed

5. **Run tests**
   - `go test ./...`

6. **Commit and open a PR**
   - `git add CHANGELOG.md`
   - `git commit -m "chore: prepare vX.Y.Z"`
   - Push the branch and open a PR: `prepare-release/vX.Y.Z` → `main`

7. **Merge the PR into `main`**
   - Prefer **squash merge** for a clean history, unless you have a different convention.

## Tag and publish

8. **Create an annotated tag from updated `main`**
   - `git checkout main && git pull`
   - `git tag -a vX.Y.Z -m "vX.Y.Z"`
   - `git push origin vX.Y.Z`

9. **Publish the GitHub Release**
   - In GitHub: Releases → Draft a new release → choose the tag `vX.Y.Z` → **Publish**.
   - Publishing triggers the release workflow, which:
     - builds and uploads artifacts
     - stamps version metadata via `-ldflags`
     - fills the release body from `CHANGELOG.md`

10. **Verify**
   - Download an artifact for your OS and run:
     - `./mochi-sticky --version`

## Hotfixes (optional)

For urgent fixes after a release:

- Create `hotfix/vX.Y.(Z+1)` from `main`, open a PR, merge, then tag and publish the patch release.

## Where changes should go

- Release notes content: `CHANGELOG.md` (the release workflow reads it).
- Version string in binaries: injected at build time (see `release/versioning`).
- Build/upload logic: `.github/workflows/release.yml` (change rarely; review carefully).
