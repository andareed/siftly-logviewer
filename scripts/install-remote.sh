#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-"$ROOT_DIR/dist"}"
DEST_ROOT="${DEST_ROOT:-/mnt/support/tools/sup/tools}"
ARCH="${ARCH:-amd64}"
SSH_TARGET="${SSH_TARGET:-csgate-via-nascent}"
SSH_CONFIG="${SSH_CONFIG:-"$HOME/.ssh/config"}"
DRY_RUN=0

SSH_OPTIONS=(
  -F "$SSH_CONFIG"
  -o RemoteCommand=none
  -o RequestTTY=no
)

usage() {
  cat <<'USAGE'
Usage: bash scripts/install-remote.sh [--dry-run] [DEST_ROOT]

Deploys the versioned Linux binaries from dist/ through SCP:
  hostlog  -> DEST_ROOT/hostlog/scripts/hostlog
  todaylog -> DEST_ROOT/todaylog/scripts/todaylog

Both remote targets are checked before upload and updated in place with
checksum verification and rollback protection. Existing target ownership,
permissions, ACLs, and security metadata are not changed.

Environment:
  DIST_DIR    Source directory, default: ./dist
  DEST_ROOT   Remote install root, default: /mnt/support/tools/sup/tools
  ARCH        Linux architecture suffix, default: amd64
  SSH_TARGET  SSH config alias, default: csgate-via-nascent
  SSH_CONFIG  SSH config file, default: ~/.ssh/config
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
if [ ! -f "$SSH_CONFIG" ]; then
  echo "SSH config not found: $SSH_CONFIG" >&2
  exit 1
fi

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
      echo "multiple Linux binaries found for $tool; clean dist/ first:" >&2
      printf '  %s\n' "${matches[@]}" >&2
      return 1
      ;;
  esac
}

binary_version() {
  local tool="$1"
  local path="$2"
  local name
  local version

  name="$(basename "$path")"
  version="${name#"${tool}_"}"
  version="${version%"_linux_${ARCH}"}"
  if [ -z "$version" ] || [ "$version" = "$name" ]; then
    echo "cannot determine version from binary name: $name" >&2
    return 1
  fi
  printf '%s\n' "$version"
}

remote_bash() {
  local script="$1"
  shift

  local command
  local arg
  local quoted

  printf -v command 'bash -c %q --' "$script"
  for arg in "$@"; do
    printf -v quoted ' %q' "$arg"
    command+="$quoted"
  done

  ssh "${SSH_OPTIONS[@]}" "$SSH_TARGET" "$command"
}

REMOTE_VALIDATE_SCRIPT=$(cat <<'REMOTE'
set -euo pipefail

for dst in "$@"; do
  if [ ! -e "$dst" ]; then
    echo "remote target does not exist: $dst" >&2
    exit 1
  fi
  if [ -L "$dst" ] || [ ! -f "$dst" ]; then
    echo "remote target is not a regular file: $dst" >&2
    exit 1
  fi
  if [ ! -w "$dst" ] || [ ! -w "$(dirname "$dst")" ]; then
    echo "remote target or directory is not writable: $dst" >&2
    exit 1
  fi
  stat -c "Remote target: %n mode=%a owner=%U:%G size=%s" "$dst"
  "$dst" --version
done
REMOTE
)

REMOTE_CLEANUP_SCRIPT=$(cat <<'REMOTE'
set -euo pipefail
rm -f -- "$@"
REMOTE
)

REMOTE_ACTIVATE_SCRIPT=$(cat <<'REMOTE'
set -euo pipefail

host_tmp="$1"
host_dst="$2"
host_sha="$3"
host_backup="$4"
today_tmp="$5"
today_dst="$6"
today_sha="$7"
today_backup="$8"

committed=0
host_metadata="$(stat -c "%a:%u:%g:%i" "$host_dst")"
today_metadata="$(stat -c "%a:%u:%g:%i" "$today_dst")"

rollback() {
  local status=$?
  trap - EXIT HUP INT TERM

  if [ "$committed" -eq 0 ]; then
    if [ -f "$host_backup" ]; then
      cat -- "$host_backup" > "$host_dst"
    fi
    if [ -f "$today_backup" ]; then
      cat -- "$today_backup" > "$today_dst"
    fi
  fi
  rm -f -- "$host_tmp" "$today_tmp" "$host_backup" "$today_backup"
  exit "$status"
}
trap rollback EXIT HUP INT TERM

prepare_upload() {
  local tmp="$1"
  local dst="$2"
  local expected_sha="$3"
  local actual_sha

  if [ ! -f "$tmp" ] || [ -L "$tmp" ]; then
    echo "uploaded file is missing or invalid: $tmp" >&2
    return 1
  fi
  if [ ! -f "$dst" ] || [ -L "$dst" ]; then
    echo "remote target changed during deployment: $dst" >&2
    return 1
  fi

  actual_sha="$(sha256sum "$tmp")"
  actual_sha="${actual_sha%% *}"
  if [ "$actual_sha" != "$expected_sha" ]; then
    echo "checksum mismatch for uploaded file: $tmp" >&2
    return 1
  fi

}

prepare_upload "$host_tmp" "$host_dst" "$host_sha"
prepare_upload "$today_tmp" "$today_dst" "$today_sha"

cp -- "$host_dst" "$host_backup"
cp -- "$today_dst" "$today_backup"

cat -- "$host_tmp" > "$host_dst"
cat -- "$today_tmp" > "$today_dst"

installed_host_sha="$(sha256sum "$host_dst")"
installed_host_sha="${installed_host_sha%% *}"
installed_today_sha="$(sha256sum "$today_dst")"
installed_today_sha="${installed_today_sha%% *}"

if [ "$installed_host_sha" != "$host_sha" ] ||
   [ "$installed_today_sha" != "$today_sha" ]; then
  echo "post-install checksum verification failed" >&2
  exit 1
fi
if [ "$(stat -c "%a:%u:%g:%i" "$host_dst")" != "$host_metadata" ] ||
   [ "$(stat -c "%a:%u:%g:%i" "$today_dst")" != "$today_metadata" ]; then
  echo "target metadata changed during deployment" >&2
  exit 1
fi

echo "Installed:"
stat -c "  %n mode=%a owner=%U:%G size=%s" "$host_dst" "$today_dst"
"$host_dst" --version
"$today_dst" --version

committed=1
rm -f -- "$host_tmp" "$today_tmp" "$host_backup" "$today_backup"
trap - EXIT HUP INT TERM
REMOTE
)

host_src="$(find_linux_binary hostlog)"
today_src="$(find_linux_binary todaylog)"
host_version="$(binary_version hostlog "$host_src")"
today_version="$(binary_version todaylog "$today_src")"

if [ "$host_version" != "$today_version" ]; then
  echo "release versions do not match: hostlog=$host_version todaylog=$today_version" >&2
  exit 1
fi

host_sha="$(sha256sum "$host_src")"
host_sha="${host_sha%% *}"
today_sha="$(sha256sum "$today_src")"
today_sha="${today_sha%% *}"

host_dst="$DEST_ROOT/hostlog/scripts/hostlog"
today_dst="$DEST_ROOT/todaylog/scripts/todaylog"
deploy_id="$(date +%Y%m%d%H%M%S)-$$"
host_tmp="$(dirname "$host_dst")/.hostlog.siftly-upload-$deploy_id"
today_tmp="$(dirname "$today_dst")/.todaylog.siftly-upload-$deploy_id"
host_backup="$(dirname "$host_dst")/.hostlog.siftly-backup-$deploy_id"
today_backup="$(dirname "$today_dst")/.todaylog.siftly-backup-$deploy_id"

echo "Release: $host_version"
echo "SSH target: $SSH_TARGET"
echo "Checking remote deployment targets..."
remote_bash "$REMOTE_VALIDATE_SCRIPT" "$host_dst" "$today_dst"

printf 'hostlog:  %s -> %s  sha256=%s\n' "$host_src" "$host_dst" "$host_sha"
printf 'todaylog: %s -> %s  sha256=%s\n' "$today_src" "$today_dst" "$today_sha"

if [ "$DRY_RUN" -eq 1 ]; then
  echo "DRY-RUN: remote targets validated; no files uploaded or changed"
  exit 0
fi

uploads_pending=1
cleanup_uploads() {
  local status=$?
  trap - EXIT HUP INT TERM
  if [ "$uploads_pending" -eq 1 ]; then
    remote_bash "$REMOTE_CLEANUP_SCRIPT" \
      "$host_tmp" "$today_tmp" "$host_backup" "$today_backup" || true
  fi
  exit "$status"
}
trap cleanup_uploads EXIT HUP INT TERM

echo "Uploading staged binaries..."
scp "${SSH_OPTIONS[@]}" "$host_src" "$SSH_TARGET:$host_tmp"
scp "${SSH_OPTIONS[@]}" "$today_src" "$SSH_TARGET:$today_tmp"

echo "Verifying and activating both binaries..."
remote_bash "$REMOTE_ACTIVATE_SCRIPT" \
  "$host_tmp" "$host_dst" "$host_sha" "$host_backup" \
  "$today_tmp" "$today_dst" "$today_sha" "$today_backup"

uploads_pending=0
trap - EXIT HUP INT TERM
echo "Deployment complete: $host_version"
