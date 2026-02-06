---
title: OSS Roadmap
slug: roadmap/roadmap-oss
section: Roadmap
order: 0
tags: []
status: published
---
What the boards say (release relevance)

default: has the OSS-release blockers (T-000015 repo essentials, T-000009 CI, T-000001 multi-platform release, T-000017 versioning/changelog, T-000018 quality hardening).
wiki-improvement: mostly “nice-to-have after first public release” (bigger wiki search/TUI search, TOC, wiki versioning/diffs).
board-improvement / adr-improvement: product expansions (exports/metrics/graphs) → defer unless you need them for launch positioning.
wiki-roadmap (archived): earlier wiki milestones are already done.
Gaps to close before going public (as of 2026-02-05)

Missing root files: LICENSE, CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md, SUPPORT.md, CHANGELOG.md (tracked by T-000015 / T-000017 / T-000018).
No GitHub workflows yet (tracked by T-000009 / T-000001).
README.md still has placeholder links (github.com/yourusername/...) → must be replaced or pointed at in-repo docs.
Roadmap to first OSS release (suggest v0.1.0, then iterate)

Milestone A — Repo + governance (1–2 days): complete T-000015, fix README placeholders, add “support policy” + “security reporting”.
Milestone B — CI quality gate (1–2 days): complete T-000009 + linter config from T-000018 (golangci-lint, go test ./..., coverage artifact).
Milestone C — Versioning contract (0.5–1 day): complete T-000017 (mochi-sticky version + --version, build stamping, lightweight CHANGELOG.md, SemVer policy: “0.x may change formats”).
Milestone D — Release automation (1–2 days): complete T-000001 (GitHub Release artifacts for linux/darwin/windows; checksums; smoke test step).
Milestone E — “First-user” docs (0.5–1 day): tighten install paths (go install ...@latest + binary download), add 1 quickstart flow + 1 screenshot/GIF (from T-000018 doc section).
After v0.1.0: pick one “headline” from each board (e.g., wiki-improvement search/TOC, board-improvement JSON/CSV export, default backup/restore) and ship v0.2.x increments.
If you tell me what you want the very first tag to be (v0.1.0 vs v1.0.0) and where docs should live (external wiki vs built-in wiki only), I can turn the roadmap into a concrete ordered task list (and even pre-fill task bodies via MCP).
