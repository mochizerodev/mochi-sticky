#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Install mochi-sticky from GitHub release artifacts.

Usage:
  install.sh [--version <tag|latest>] [--repo <owner/repo>] [--install-dir <dir>]
             [--from-file <archive>] [--checksum-file <file>] [--skip-checksum] [--force]

Examples:
  curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | bash
  curl -fsSL https://raw.githubusercontent.com/mochizerodev/mochi-sticky/main/scripts/install.sh | \
    bash -s -- --version v0.6.0
EOF
}

VERSION="latest"
REPO="mochizerodev/mochi-sticky"
INSTALL_DIR=""
FROM_FILE=""
CHECKSUM_FILE=""
SKIP_CHECKSUM=0
FORCE=0
BIN_NAME="mochi-sticky"

while (($# > 0)); do
  case "$1" in
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --repo)
      REPO="${2:-}"
      shift 2
      ;;
    --install-dir)
      INSTALL_DIR="${2:-}"
      shift 2
      ;;
    --from-file)
      FROM_FILE="${2:-}"
      shift 2
      ;;
    --checksum-file)
      CHECKSUM_FILE="${2:-}"
      shift 2
      ;;
    --skip-checksum)
      SKIP_CHECKSUM=1
      shift
      ;;
    --force)
      FORCE=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

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

verify_checksum() {
  local archive="$1"
  local checksum_path="$2"
  local expected actual
  expected="$(awk '{print $1}' "$checksum_path" | head -n 1 | tr '[:upper:]' '[:lower:]')"
  if [[ -z "$expected" ]]; then
    echo "Checksum file is empty: $checksum_path" >&2
    return 1
  fi
  actual="$(hash_file "$archive" | tr '[:upper:]' '[:lower:]')"
  if [[ "$expected" != "$actual" ]]; then
    echo "Checksum mismatch for $archive" >&2
    echo "Expected: $expected" >&2
    echo "Actual:   $actual" >&2
    return 1
  fi
}

detect_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *)
      echo "Unsupported OS: $os (expected Linux or Darwin)" >&2
      return 1
      ;;
  esac

  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)
      echo "Unsupported architecture: $arch (expected amd64 or arm64)" >&2
      return 1
      ;;
  esac

  printf '%s %s\n' "$os" "$arch"
}

resolve_install_dir() {
  local target="${INSTALL_DIR}"
  if [[ -n "$target" ]]; then
    echo "$target"
    return 0
  fi
  if [[ "$(id -u)" -eq 0 || -w "/usr/local/bin" ]]; then
    echo "/usr/local/bin"
    return 0
  fi
  echo "${HOME}/.local/bin"
}

main() {
  local archive_path checksum_path target_dir install_path os arch base_name archive_name base_url extracted_path
  tmp_dir="$(mktemp -d)"
  trap '[[ -n "${tmp_dir:-}" ]] && rm -rf "${tmp_dir}"' EXIT

  if [[ -n "$FROM_FILE" ]]; then
    archive_path="$FROM_FILE"
    if [[ ! -f "$archive_path" ]]; then
      echo "Archive not found: $archive_path" >&2
      exit 1
    fi
  else
    if ! command -v curl >/dev/null 2>&1; then
      echo "curl is required for remote downloads." >&2
      exit 1
    fi
    read -r os arch <<<"$(detect_platform)"
    base_name="${BIN_NAME}-${os}-${arch}"
    archive_name="${base_name}.tar.gz"
    if [[ "$VERSION" == "latest" ]]; then
      base_url="https://github.com/${REPO}/releases/latest/download"
    else
      base_url="https://github.com/${REPO}/releases/download/${VERSION}"
    fi
    archive_path="${tmp_dir}/${archive_name}"
    curl -fsSL "${base_url}/${archive_name}" -o "${archive_path}"
    if [[ -z "$CHECKSUM_FILE" ]]; then
      checksum_path="${tmp_dir}/${archive_name}.sha256"
      curl -fsSL "${base_url}/${archive_name}.sha256" -o "${checksum_path}"
      CHECKSUM_FILE="${checksum_path}"
    fi
  fi

  if [[ "$SKIP_CHECKSUM" -eq 0 ]]; then
    if [[ -z "$CHECKSUM_FILE" ]]; then
      if [[ -f "${archive_path}.sha256" ]]; then
        CHECKSUM_FILE="${archive_path}.sha256"
      else
        echo "Checksum file is required (use --checksum-file or --skip-checksum)." >&2
        exit 1
      fi
    fi
    verify_checksum "$archive_path" "$CHECKSUM_FILE"
  fi

  tar -xzf "$archive_path" -C "$tmp_dir"
  extracted_path="${tmp_dir}/${BIN_NAME}"
  if [[ ! -f "$extracted_path" ]]; then
    extracted_path="$(find "${tmp_dir}" -maxdepth 2 -type f -name "${BIN_NAME}" | head -n 1 || true)"
  fi
  if [[ -z "$extracted_path" || ! -f "$extracted_path" ]]; then
    echo "Failed to find ${BIN_NAME} in archive ${archive_path}" >&2
    exit 1
  fi

  target_dir="$(resolve_install_dir)"
  mkdir -p "$target_dir"
  install_path="${target_dir}/${BIN_NAME}"
  if [[ -e "$install_path" && "$FORCE" -eq 0 ]]; then
    echo "Binary already exists at ${install_path}. Re-run with --force to replace." >&2
    exit 1
  fi
  install -m 755 "$extracted_path" "$install_path"

  echo "Installed ${BIN_NAME} to ${install_path}"
  if [[ ":${PATH}:" != *":${target_dir}:"* ]]; then
    echo "Add ${target_dir} to PATH to run ${BIN_NAME} from any shell."
  fi
  echo "Verify with: ${BIN_NAME} --version"
}

main
