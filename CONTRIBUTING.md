# Contributing to mochi-sticky

Thanks for your interest in contributing! This guide explains how to get set up and submit changes.

## Development setup

Prerequisites:
- Go 1.23.12+ (see `go.mod`)
- Git

Clone and build:
```bash
git clone https://github.com/mochizero0/mochi-sticky.git
cd mochi-sticky
go build ./...
```

Run tests:
```bash
go test ./...
go vet ./...
```

Optional linting (if installed):
```bash
golangci-lint run
```

Optional pre-commit hook (recommended):
```bash
pre-commit install
```

Install `pre-commit` (choose one):
```bash
# macOS (Homebrew)
brew install pre-commit

# Ubuntu/Debian
sudo apt-get install pre-commit

# Python (any OS)
python -m pip install --user pre-commit
```

Run hooks on all files:
```bash
pre-commit run --all-files
```

The hook runs `golangci-lint` before each commit.

## Project structure

See `MASTER_SPEC.md` and the built-in wiki (`mochi-sticky wiki view architecture`) for architecture and coding standards.

## AI Assistance & License Hygiene

AI-assisted tooling is welcome for drafting and refactoring, but:
- Maintainers review all changes before merge (treat AI output like any other contribution).
- Don’t paste in code you don’t have rights to (license compatibility).
- Don’t include secrets or private data in prompts.

## Making changes

1. Fork the repo and create a feature branch.
2. Keep changes focused and update docs if behavior changes.
3. Add or update tests where appropriate.
4. Ensure tests pass before opening a PR.

## Commit messages

We recommend Conventional Commits (e.g., `feat: add task search filter`, `fix: handle empty board config`) but do not enforce them.

## Pull requests

Include:
- A concise summary of the change and motivation.
- Tests or reproduction steps.
- Screenshots/GIFs for UI/TUI changes, if relevant.

## Reporting issues

Please use GitHub Issues for bugs and feature requests. Security issues should follow `SECURITY.md`.
