---
title: Homebrew
slug: release/homebrew
section: Release
order: 6
tags:
    - release
    - homebrew
    - install
status: published
---
# Homebrew distribution (tap)

Homebrew is an optional distribution channel for `mochi-sticky`. You do not need to "register" with Homebrew. You publish a formula either:

- In your own tap (recommended to start): a GitHub repo that contains `Formula/mochi-sticky.rb`.
- In `homebrew/core` (best long-term): a PR to Homebrew's core repo.

This page documents the tap approach for `mochizerodev/mochi-sticky`.

## What Homebrew installs (and what it doesn't)

- Homebrew installs the `mochi-sticky` binary.
- The release installer scripts (`scripts/install.sh`, `scripts/install.ps1`) are not required for Homebrew installs. Keep using them for users who prefer `curl | bash` or PowerShell install flows.

## 1. Create a tap repository

Create a public GitHub repo:

- Name: `homebrew-tap` (convention)
- Example: `mochizerodev/homebrew-tap`

Repo structure:

```text
homebrew-tap/
  Formula/
    mochi-sticky.rb
  README.md
```

## 2. Formula basics

A formula needs:

- `url`: points to a release artifact
- `sha256`: checksum of that artifact
- `version`: tag version (optional if baked into URL)
- `install`: copies the binary into Homebrew's prefix
- `test`: a simple runtime sanity check

For `mochi-sticky`, prefer using the packaged release archives we publish:

- macOS: `mochi-sticky-darwin-amd64.tar.gz` / `mochi-sticky-darwin-arm64.tar.gz`
- Linux: `mochi-sticky-linux-amd64.tar.gz` / `mochi-sticky-linux-arm64.tar.gz`

Each archive has a sibling checksum file: `<archive>.sha256`.

## 3. Example formula (multi-OS, multi-arch)

Replace `VERSION` and the `sha256` values with your release tag and checksums.

```ruby
class MochiSticky < Formula
  desc "The Sticky-Note Project Manager for Developers"
  homepage "https://github.com/mochizerodev/mochi-sticky"
  version "VERSION"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/mochizerodev/mochi-sticky/releases/download/vVERSION/mochi-sticky-darwin-arm64.tar.gz"
      sha256 "DARWIN_ARM64_SHA256"
    else
      url "https://github.com/mochizerodev/mochi-sticky/releases/download/vVERSION/mochi-sticky-darwin-amd64.tar.gz"
      sha256 "DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/mochizerodev/mochi-sticky/releases/download/vVERSION/mochi-sticky-linux-arm64.tar.gz"
      sha256 "LINUX_ARM64_SHA256"
    else
      url "https://github.com/mochizerodev/mochi-sticky/releases/download/vVERSION/mochi-sticky-linux-amd64.tar.gz"
      sha256 "LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "mochi-sticky"
  end

  test do
    assert_match "mochi-sticky", shell_output("#{bin}/mochi-sticky --version")
  end
end
```

Notes:

- Homebrew runs `install` with the archive already downloaded and extracted.
- The archive must contain a `mochi-sticky` binary at the root of the extracted directory.

## 4. Publishing and using the tap

In the tap repo:

1. Commit `Formula/mochi-sticky.rb`.
2. Create a release/tag for the tap repo if you want, but it is not required.

Users install via:

```bash
brew tap mochizerodev/tap
brew install mochi-sticky
mochi-sticky --version
```

If your tap repo is named `homebrew-tap`, Homebrew conventionally uses:

```bash
brew tap mochizerodev/tap
```

(Where `tap` is the repo name without the `homebrew-` prefix.)

## 5. Updating the tap for each new release

When a new `mochi-sticky` release is published (for example `v0.6.0`), update the tap repo formula at `mochizerodev/homebrew-tap`.

1. Sync the tap repo locally.

```bash
git clone git@github.com:mochizerodev/homebrew-tap.git
cd homebrew-tap
git checkout main
git pull
```

2. Fetch checksums for the new release assets.

```bash
TAG=v0.6.0
BASE="https://github.com/mochizerodev/mochi-sticky/releases/download/${TAG}"

curl -fsSL "${BASE}/mochi-sticky-darwin-arm64.tar.gz.sha256" | awk '{print $1}'
curl -fsSL "${BASE}/mochi-sticky-darwin-amd64.tar.gz.sha256" | awk '{print $1}'
curl -fsSL "${BASE}/mochi-sticky-linux-arm64.tar.gz.sha256" | awk '{print $1}'
curl -fsSL "${BASE}/mochi-sticky-linux-amd64.tar.gz.sha256" | awk '{print $1}'
```

3. Edit `Formula/mochi-sticky.rb`.

- Set `version "0.6.0"` (no leading `v`).
- Update each `url` to `.../download/v0.6.0/...`.
- Replace all four `sha256` values with the checksums from step 2.

4. Validate the formula locally.

```bash
brew install ./Formula/mochi-sticky.rb
brew test mochi-sticky
mochi-sticky --version
```

5. Commit and push the tap update.

```bash
git add Formula/mochi-sticky.rb
git commit -m "mochi-sticky v0.6.0"
git push origin main
```

6. After the push, users get the new release with:

```bash
brew update
brew upgrade mochi-sticky
```

Short alternative (when local tap state is stale):

```bash
brew untap mochizerodev/tap
brew tap mochizerodev/tap
brew upgrade mochi-sticky
```

## 6. Automating tap updates (recommended)

Recommended automation approach:

- Add a GitHub Actions workflow in the tap repo that runs on a schedule or is triggered manually.
- It fetches the latest `mochi-sticky` release tag and checksums, rewrites the formula, and opens a PR.

If you prefer a push-based approach, you can also have the `mochi-sticky` release workflow trigger the tap repo workflow using a repository dispatch event or a GitHub App token.

Implementation details depend on how you manage secrets and whether the tap repo is in the same org.

## 7. Relationship to the installer scripts

You now have three install paths:

- Homebrew (`brew install mochi-sticky`): best for Homebrew users.
- Shell/PowerShell installers (`scripts/install.sh`, `scripts/install.ps1`): best for non-Homebrew installs and CI.
- Direct download: users can grab `*.tar.gz` / `*.zip` + `.sha256` from GitHub Releases.

For release packaging details, see `release/distribution`.
