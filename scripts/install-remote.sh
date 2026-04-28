#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-"$ROOT_DIR/dist"}"
DEST_ROOT="${DEST_ROOT:-/mnt/support/tools/sup/tools}"
ARCH="${ARCH:-amd64}"
DRY_RUN=0

usage() {
  cat <<'USAGE'
Usage: scripts/install-remote.sh [--dry-run] [DEST_ROOT]

Installs versioned Linux binaries from dist/ into the remote tool layout:
  hostlog  -> DEST_ROOT/hostlog/scripts/hostlog
  todaylog -> DEST_ROOT/todaylog/scripts/todaylog
  devfmt   -> DEST_ROOT/devinfo/scripts/devinfo

Environment:
  DIST_DIR   Source directory, default: ./dist
  DEST_ROOT  Install root, default: /mnt/support/tools/sup/tools
  ARCH       Linux arch suffix, default: amd64
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run|-n)
      DRY_RUN=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    -*)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      DEST_ROOT="$1"
      shift
      ;;
  esac
done

if [ ! -d "$DIST_DIR" ]; then
  echo "dist directory not found: $DIST_DIR" >&2
  exit 1
fi

if [ ! -d "$DEST_ROOT" ]; then
  echo "destination root not found: $DEST_ROOT" >&2
  exit 1
fi

run() {
  if [ "$DRY_RUN" -eq 1 ]; then
    printf 'DRY-RUN:'
    printf ' %q' "$@"
    printf '\n'
    return 0
  fi
  "$@"
}

find_linux_binary() {
  local tool="$1"
  local matches=()

  shopt -s nullglob
  matches=("$DIST_DIR"/"${tool}"_*"_linux_${ARCH}")
  shopt -u nullglob

  case "${#matches[@]}" in
    0)
      echo "no Linux binary found for $tool matching: $DIST_DIR/${tool}_*_linux_${ARCH}" >&2
      return 1
      ;;
    1)
      printf '%s\n' "${matches[0]}"
      ;;
    *)
      echo "multiple Linux binaries found for $tool; set ARCH or clean dist/:" >&2
      printf '  %s\n' "${matches[@]}" >&2
      return 1
      ;;
  esac
}

install_tool() {
  local source_tool="$1"
  local target_tool_dir="$2"
  local target_binary="$3"
  local src
  local dst_dir
  local dst

  src="$(find_linux_binary "$source_tool")"
  dst_dir="$DEST_ROOT/$target_tool_dir/scripts"
  dst="$dst_dir/$target_binary"

  if [ ! -d "$dst_dir" ]; then
    echo "destination scripts directory not found: $dst_dir" >&2
    return 1
  fi

  echo "Installing $(basename "$src") -> $dst"
  run install -m 700 "$src" "$dst"
}

install_tool hostlog hostlog hostlog
install_tool todaylog todaylog todaylog
install_tool devfmt devinfo devinfo
