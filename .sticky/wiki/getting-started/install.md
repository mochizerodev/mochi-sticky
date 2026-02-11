---
title: Installation
slug: getting-started/install
section: Getting Started
order: 1
tags:
    - install
status: published
---
# Installation

This guide covers release installers (Linux/macOS/Windows), upgrade/downgrade flows, uninstall paths, and source builds.

## Prerequisites

- Linux/macOS: `curl`, `tar`, and `sha256sum` (or `shasum`)
- Windows: PowerShell 5+ and `Invoke-WebRequest`
- Optional for source builds: Go 1.24.6+

## 1. Fresh install (release artifacts)

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | bash
```

The installer chooses `/usr/local/bin` when writable; otherwise it installs to `~/.local/bin`.

### macOS

Primary path today is the same shell installer:

```bash
curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | bash
```

Optional Homebrew path (if a tap/formula is published for your target tag):

```bash
brew tap <tap-owner>/<tap-name>
brew install mochi-sticky
```

### Windows

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.ps1" -OutFile "install-mochi-sticky.ps1"
powershell -ExecutionPolicy Bypass -File .\install-mochi-sticky.ps1
```

By default, this installs to `%LOCALAPPDATA%\Programs\mochi-sticky\bin`.

## 2. Verify install

Linux/macOS:

```bash
mochi-sticky --version
tmpdir="$(mktemp -d)"
(cd "$tmpdir" && mochi-sticky init && mochi-sticky board list)
```

Windows:

```powershell
mochi-sticky --version
$tmp = Join-Path $env:TEMP ("mochi-sticky-" + [Guid]::NewGuid().ToString("N"))
New-Item -Path $tmp -ItemType Directory | Out-Null
Push-Location $tmp
mochi-sticky init
mochi-sticky board list
Pop-Location
```

## 3. Upgrade or downgrade

### Linux/macOS

Upgrade to latest:

```bash
curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | bash -s -- --force
```

Install a specific version (upgrade or downgrade):

```bash
curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | bash -s -- --version v0.6.0 --force
```

### Windows

Upgrade to latest:

```powershell
powershell -ExecutionPolicy Bypass -File .\install-mochi-sticky.ps1 -Force
```

Install a specific version (upgrade or downgrade):

```powershell
powershell -ExecutionPolicy Bypass -File .\install-mochi-sticky.ps1 -Version v0.6.0 -Force
```

### Homebrew upgrade path (when using brew)

```bash
brew update
brew upgrade mochi-sticky
```

## 4. Uninstall

### Linux/macOS

If installed to `/usr/local/bin`:

```bash
sudo rm -f /usr/local/bin/mochi-sticky
```

If installed to user-local bin:

```bash
rm -f ~/.local/bin/mochi-sticky
```

Optional cleanup:

```bash
rm -rf .sticky
```

### Windows

```powershell
Remove-Item "$env:LOCALAPPDATA\Programs\mochi-sticky\bin\mochi-sticky.exe" -ErrorAction SilentlyContinue
```

Optional cleanup:

```powershell
Remove-Item ".sticky" -Recurse -Force -ErrorAction SilentlyContinue
```

### Homebrew uninstall path (when using brew)

```bash
brew uninstall mochi-sticky
```

## Initialize Your First Project

Navigate to your project directory and initialize mochi-sticky:

```bash
cd /path/to/your/project
mochi-sticky init
```

This creates:
- `.sticky/` — Data directory for boards, tasks, and configuration
- `.sticky/boards/boards.yaml` — Board registry
- `.sticky/boards/default/` — Default board with initial configuration
- `.sticky/templates/` — Default templates (task, board, ADR, wiki)
- `.sticky/wiki/` — Wiki documentation root
- `.sticky/adrs/` — Architecture Decision Records

**Note**: `mochi-sticky tui --set-editor "code --wait"` writes the editor preference to `.sticky/mochi-sticky.yaml` (`editor`).

## Storage Root Configuration

By default, mochi-sticky uses `.sticky/` in your current directory. You can customize this:

### Method 1: Configuration File

Create `.sticky/mochi-sticky.yaml`:

```yaml
storage_root: .sticky
editor: "code --wait"
config_paths:
  boards: boards/boards.yaml
  adr: adrs/config.yaml
  wiki_index: wiki/_index.yaml
templates:
  root: .sticky/templates
  adr: .sticky/templates/adr
  task: .sticky/templates/task
  board: .sticky/templates/board
  wiki: .sticky/templates/wiki
  wiki_pdf: .sticky/templates/wiki/wiki_pdf_template.tex
```

If you already have a legacy `mochi-sticky.yaml` in the project root, it is still honored for `storage_root`.

### Method 2: Environment Variable

```bash
export MOCHI_STICKY_STORAGE=/path/to/custom/storage
mochi-sticky init
```

### Method 3: CLI Flag

```bash
mochi-sticky --storage /path/to/custom/storage init
```

**Priority order:** CLI flag > Environment variable > `.sticky/mochi-sticky.yaml` > legacy root `mochi-sticky.yaml` > Default (`.sticky/`)

## Validate Installation

Check that everything is configured correctly:

```bash
mochi-sticky hydrate
```

This outputs:
- Storage root location
- Board registry path
- Active board
- Template directories
- Configuration validity

For JSON output (useful for automation):
```bash
mochi-sticky hydrate --json --pretty
```

## Troubleshooting

### "command not found: mochi-sticky"

Ensure the binary is in your `$PATH`:
```bash
echo $PATH
which mochi-sticky
```

### PATH includes install directory but command still fails

Open a new shell (or restart PowerShell) so the updated PATH is loaded, then run:

```bash
mochi-sticky --version
```

### "permission denied"

Install to a user-owned directory:

```bash
curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | \
  bash -s -- --install-dir "$HOME/.local/bin" --force
```

If you expect a system install, re-run with elevated permissions:

```bash
sudo bash -c 'curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | bash -s -- --force'
```

### Checksum mismatch during install

1. Retry download/installation (corrupt partial downloads are the most common cause).
2. Verify you are using matching artifact and `.sha256` files from the same release tag.
3. Avoid caching/proxy rewriting if your environment mirrors release artifacts.

### Build Errors (source install only)

Ensure you have Go 1.24.6+ installed:

```bash
go version
```

Update dependencies:

```bash
go mod tidy
go mod download
```

## 5. Build from source (alternative path)

```bash
git clone https://github.com/mochizerodev/mochi-sticky.git
cd mochi-sticky
go build -o mochi-sticky
```

Install manually:

Linux/macOS:

```bash
sudo cp mochi-sticky /usr/local/bin/
# or:
mkdir -p ~/.local/bin
cp mochi-sticky ~/.local/bin/
```

Windows:

```powershell
go build -o mochi-sticky.exe
Copy-Item .\mochi-sticky.exe "$env:LOCALAPPDATA\Programs\mochi-sticky\bin\mochi-sticky.exe" -Force
```
