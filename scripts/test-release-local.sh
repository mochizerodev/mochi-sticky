#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Local test harness for the GitHub Actions release workflow (build + smoke), using local artifacts.

Usage:
  bash scripts/test-release-local.sh [--version <tag>] [--goos <os[,os...]>] [--goarch <arch[,arch...]>]
                                     [--dist-dir <dir>] [--skip-tests] [--no-smoke] [--keep-temp]
                                     [--offline]

Defaults:
  --version  v0.0.0-local
  --goos     host OS only
  --goarch   host arch only
  --dist-dir dist
  smoke      enabled (host target only)

Notes:
  - Windows smoke testing requires running on Windows. This script will still build Windows artifacts, but will skip smoke on non-Windows hosts.
  - Uses scripts/install.sh for unix smoke (Linux/macOS), and expects archives to contain a generic binary name (mochi-sticky / mochi-sticky.exe).
  - --offline disables module/toolchain downloads. It attempts to seed the temp module cache from your system module cache.
EOF
}

VERSION="v0.0.0-local"
GOOS_LIST=""
GOARCH_LIST=""
DIST_DIR="dist"
SKIP_TESTS=0
NO_SMOKE=0
KEEP_TEMP=0
OFFLINE=0
GO_BIN="go"

while (($# > 0)); do
  case "$1" in
    --version) VERSION="${2:-}"; shift 2 ;;
    --goos) GOOS_LIST="${2:-}"; shift 2 ;;
    --goarch) GOARCH_LIST="${2:-}"; shift 2 ;;
    --dist-dir) DIST_DIR="${2:-}"; shift 2 ;;
    --skip-tests) SKIP_TESTS=1; shift ;;
    --no-smoke) NO_SMOKE=1; shift ;;
    --keep-temp) KEEP_TEMP=1; shift ;;
    --offline) OFFLINE=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

repo_root() {
  local root
  root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  printf '%s\n' "$root"
}

hash_file() {
  local target="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$target" | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$target" | awk '{print $1}'
    return 0
  fi
  echo "No SHA256 tool found (sha256sum or shasum)." >&2
  return 1
}

host_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *) os="unknown" ;;
  esac

  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) arch="unknown" ;;
  esac

  printf '%s %s\n' "$os" "$arch"
}

split_csv() {
  local csv="$1"
  if [[ -z "$csv" ]]; then
    return 0
  fi
  local IFS=','
  read -r -a parts <<<"$csv"
  printf '%s\n' "${parts[@]}"
}

git_commit() {
  if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git rev-parse HEAD
    return 0
  fi
  echo "unknown"
}

dir_writable() {
  local dir="$1"
  if [[ -z "$dir" || ! -d "$dir" ]]; then
    return 1
  fi
  local probe="${dir}/.write-probe.$RANDOM.$RANDOM"
  if ( : > "$probe" ) 2>/dev/null; then
    rm -f "$probe" 2>/dev/null || true
    return 0
  fi
  return 1
}

build_one() {
  local goos="$1"
  local goarch="$2"
  local dist="$3"
  local commit date ldflags base asset_bin internal_bin archive checksum meta stage

  commit="$(git_commit)"
  date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  ldflags="-X mochi-sticky/internal/version.Version=${VERSION} -X mochi-sticky/internal/version.Commit=${commit} -X mochi-sticky/internal/version.Date=${date}"

  base="mochi-sticky-${goos}-${goarch}"
  asset_bin="${base}"
  internal_bin="mochi-sticky"
  archive="${base}.tar.gz"
  if [[ "$goos" == "windows" ]]; then
    asset_bin="${asset_bin}.exe"
    internal_bin="mochi-sticky.exe"
    archive="${base}.zip"
  fi
  checksum="${archive}.sha256"
  meta="${base}.metadata.json"

  mkdir -p "$dist"

  echo "==> build ${goos}/${goarch}"
  GOOS="$goos" GOARCH="$goarch" "${GO_BIN}" build -ldflags "$ldflags" -o "${dist}/${asset_bin}" .

  stage="$(mktemp -d)"

  cp "${dist}/${asset_bin}" "${stage}/${internal_bin}"

  if [[ "$goos" == "windows" ]]; then
    STAGE_DIR="$stage" INTERNAL_BIN="$internal_bin" OUT_ZIP="${dist}/${archive}" python3 - <<'PY'
import os
import zipfile

stage = os.environ["STAGE_DIR"]
internal = os.environ["INTERNAL_BIN"]
out_path = os.environ["OUT_ZIP"]
with zipfile.ZipFile(out_path, "w", compression=zipfile.ZIP_DEFLATED) as z:
    z.write(os.path.join(stage, internal), arcname=internal)
PY
  else
    tar -C "$stage" -czf "${dist}/${archive}" "$internal_bin"
  fi

  if [[ "$KEEP_TEMP" -eq 0 ]]; then
    rm -rf "$stage"
  else
    echo "==> kept staging dir: ${stage}"
  fi

  local archive_hash
  archive_hash="$(hash_file "${dist}/${archive}")"
  printf '%s  %s\n' "$archive_hash" "$archive" > "${dist}/${checksum}"

  cat > "${dist}/${meta}" <<JSON
{
  "schema_version": 1,
  "artifact": "${archive}",
  "checksum": "${checksum}",
  "platform": "${goos}",
  "arch": "${goarch}",
  "tag": "${VERSION}",
  "commit": "${commit}",
  "binary_asset": "${asset_bin}",
  "binary_in_archive": "${internal_bin}",
  "install_script_unix": "scripts/install.sh",
  "install_script_windows": "scripts/install.ps1"
}
JSON

  echo "==> wrote ${dist}/${archive} and ${dist}/${checksum}"
}

smoke_unix() {
  local root="$1"
  local dist="$2"
  local base="$3"
  local archive="${base}.tar.gz"
  local checksum="${archive}.sha256"

  echo "==> smoke (unix) ${base}"
  if [[ ! -f "${dist}/${archive}" ]]; then
    echo "Missing archive: ${dist}/${archive}" >&2
    exit 1
  fi
  if [[ ! -f "${dist}/${checksum}" ]]; then
    echo "Missing checksum: ${dist}/${checksum}" >&2
    exit 1
  fi

  local expected actual
  expected="$(awk '{print $1}' "${dist}/${checksum}" | head -n 1 | tr '[:upper:]' '[:lower:]')"
  actual="$(hash_file "${dist}/${archive}" | tr '[:upper:]' '[:lower:]')"
  if [[ "$expected" != "$actual" ]]; then
    echo "Checksum mismatch for ${dist}/${archive}" >&2
    echo "Expected: $expected" >&2
    echo "Actual:   $actual" >&2
    exit 1
  fi

  local bindir workdir
  bindir="$(mktemp -d)/bin"
  mkdir -p "$bindir"

  bash "${root}/scripts/install.sh" \
    --from-file "${dist}/${archive}" \
    --checksum-file "${dist}/${checksum}" \
    --install-dir "${bindir}" \
    --force

  PATH="${bindir}:${PATH}" mochi-sticky --version

  workdir="$(mktemp -d)"
  (
    cd "$workdir"
    PATH="${bindir}:${PATH}" mochi-sticky init
    PATH="${bindir}:${PATH}" mochi-sticky board list
  )

  bash "${root}/scripts/install.sh" \
    --from-file "${dist}/${archive}" \
    --checksum-file "${dist}/${checksum}" \
    --install-dir "${bindir}" \
    --force

  rm -f "${bindir}/mochi-sticky"
  if [[ -f "${bindir}/mochi-sticky" ]]; then
    echo "Uninstall failed: ${bindir}/mochi-sticky still exists" >&2
    exit 1
  fi

  echo "==> smoke ok"
}

main() {
  local root host_os host_arch goos_targets goarch_targets
  root="$(repo_root)"
  cd "$root"

  read -r host_os host_arch <<<"$(host_platform)"
  if [[ "$host_os" == "unknown" || "$host_arch" == "unknown" ]]; then
    echo "Unsupported host platform: ${host_os}/${host_arch}" >&2
    exit 1
  fi

  local sys_gopath sys_gomodcache sys_gocache
  sys_gopath="$(go env GOPATH 2>/dev/null || true)"
  sys_gomodcache="$(go env GOMODCACHE 2>/dev/null || true)"
  sys_gocache="$(go env GOCACHE 2>/dev/null || true)"

  if [[ "$OFFLINE" -eq 1 ]]; then
    export GOPROXY=off
    export GOSUMDB=off
    export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"

    # Force caches into writable locations for offline runs.
    export GOPATH="${GOPATH:-/tmp/mochi-sticky-gopath}"
    export GOMODCACHE="${GOMODCACHE:-${GOPATH}/pkg/mod}"
    export GOCACHE="${GOCACHE:-/tmp/mochi-sticky-go-build}"
    mkdir -p "$GOPATH" "$GOMODCACHE" "$GOCACHE"

    local seed_modcache seed_gocache
    seed_modcache="${sys_gomodcache:-${HOME}/go/pkg/mod}"
    seed_gocache="${sys_gocache:-${HOME}/.cache/go-build}"

    # Best-effort seed to avoid network access in restricted environments.
    if [[ -d "$seed_modcache" && -z "$(ls -A "$GOMODCACHE" 2>/dev/null || true)" ]]; then
      echo "==> seeding module cache from ${seed_modcache}"
      cp -R "${seed_modcache}/." "$GOMODCACHE/" 2>/dev/null || true
      chmod -R u+rwX "$GOMODCACHE" 2>/dev/null || true
    fi
    if [[ -d "$seed_gocache" && -z "$(ls -A "$GOCACHE" 2>/dev/null || true)" ]]; then
      echo "==> seeding build cache from ${seed_gocache}"
      cp -R "${seed_gocache}/." "$GOCACHE/" 2>/dev/null || true
      chmod -R u+rwX "$GOCACHE" 2>/dev/null || true
    fi

    # Prefer using a local toolchain GOROOT (if present) to avoid toolchain module verification/downloads.
    local required_go toolchain_dir
    required_go="$(awk '$1 == "go" { print $2; exit }' go.mod 2>/dev/null || true)"
    if [[ -n "$required_go" ]]; then
      toolchain_dir="${seed_modcache}/golang.org/toolchain@v0.0.1-go${required_go}.${host_os}-${host_arch}"
      if [[ -x "${toolchain_dir}/bin/go" ]]; then
        GO_BIN="${toolchain_dir}/bin/go"
      fi
    fi
    export GOTOOLCHAIN=local
  else
    # Prefer system caches when writable; fall back to /tmp otherwise (useful in sandboxed environments).
    if [[ -z "${GOPATH:-}" && -n "$sys_gopath" ]]; then
      export GOPATH="$sys_gopath"
    fi
    if [[ -z "${GOMODCACHE:-}" && -n "$sys_gomodcache" ]]; then
      export GOMODCACHE="$sys_gomodcache"
    fi
    if [[ -z "${GOCACHE:-}" && -n "$sys_gocache" ]]; then
      export GOCACHE="$sys_gocache"
    fi

    if ! dir_writable "${GOMODCACHE:-}" || ! dir_writable "${GOCACHE:-}"; then
      export GOPATH="${GOPATH:-/tmp/mochi-sticky-gopath}"
      export GOMODCACHE="${GOMODCACHE:-${GOPATH}/pkg/mod}"
      export GOCACHE="${GOCACHE:-/tmp/mochi-sticky-go-build}"
      mkdir -p "$GOPATH" "$GOMODCACHE" "$GOCACHE"
    fi
  fi

  if [[ -z "$GOOS_LIST" ]]; then
    GOOS_LIST="$host_os"
  fi
  if [[ -z "$GOARCH_LIST" ]]; then
    GOARCH_LIST="$host_arch"
  fi

  goos_targets=()
  while IFS= read -r item; do
    [[ -n "$item" ]] && goos_targets+=("$item")
  done < <(split_csv "$GOOS_LIST")

  goarch_targets=()
  while IFS= read -r item; do
    [[ -n "$item" ]] && goarch_targets+=("$item")
  done < <(split_csv "$GOARCH_LIST")

  mkdir -p "$DIST_DIR"

  if [[ "$SKIP_TESTS" -eq 0 ]]; then
    echo "==> go test ./..."
    "${GO_BIN}" test ./...
  fi

  local goos goarch
  for goos in "${goos_targets[@]}"; do
    for goarch in "${goarch_targets[@]}"; do
      build_one "$goos" "$goarch" "$DIST_DIR"
    done
  done

  if [[ "$NO_SMOKE" -eq 0 ]]; then
    if [[ "$host_os" == "windows" ]]; then
      echo "Smoke on Windows is not supported by this bash harness; run the PowerShell installer flow on Windows instead." >&2
    else
      local base="mochi-sticky-${host_os}-${host_arch}"
      smoke_unix "$root" "$DIST_DIR" "$base"
    fi
  fi

  echo "==> done"
  ls -la "$DIST_DIR" | sed -n '1,200p'
}

main
