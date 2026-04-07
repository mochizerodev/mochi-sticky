# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

- TUI v2: new app shell with tabs for Boards, Wiki, and ADRs.
- Boards: responsive Kanban view with improved resize rules.
- Boards: list view with sorting and filtering UI.
- Unified modals and toast feedback across the TUI.
- Improved wiki interaction and focus behavior.
- Hide unused options in the bottom menu.

## [v0.1.1]

- Cross-platform installers: `scripts/install.sh` (Linux/macOS) and `scripts/install.ps1` (Windows) with checksum verification, version pinning, custom install dir support, and forced replace mode for upgrades.
- Release workflow now publishes platform archives (`.tar.gz`/`.zip`), `.sha256` checksum files, and artifact metadata JSON; smoke tests install from built artifacts and validate basic CLI commands.
- Fixed release packaging status handling so archive creation failures are not masked by cleanup.
- Fixed macOS smoke checksum verification by using portable SHA256 hashing (supports `sha256sum` or `shasum`).
- Fixed `scripts/install.sh` so `curl` is only required for remote downloads (local `--from-file` installs work without it).

## [v0.1.0]

- Initial public beta release.
- Task + board management via CLI and interactive TUI.
- Built-in wiki + ADR support stored alongside your repo.
- MCP server support for AI agents and IDE integrations.
