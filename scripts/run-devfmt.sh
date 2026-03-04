#!/usr/bin/env bash
set -euo pipefail

GROUP="${1:-sw}"
export GOCACHE="${GOCACHE:-/tmp/go-build}"

zcat testdata/devinfo.dump.gz | go run ./cmd/devfmt export --debug=debug.log --input - --group "${GROUP}"
