#!/usr/bin/env bash
# Run from repo root: ./scripts/migrate.sh [goose args]
# Default: goose up (applies pending SQL migrations in database/schema).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

set -a
if [ -f .env ]; then
  # shellcheck disable=SC1091
  . ./.env
fi
set +a

if [ -z "${GOOSE_DRIVER:-}" ]; then
  echo "GOOSE_DRIVER is not set. Add it to .env (e.g. GOOSE_DRIVER=postgres)." >&2
  exit 1
fi

if [ -z "${GOOSE_DBSTRING:-}" ]; then
  echo "GOOSE_DBSTRING is not set. Add it to .env (e.g. postgres://user:pass@localhost:5432/debtease?sslmode=disable)." >&2
  exit 1
fi

export GOOSE_MIGRATION_DIR=database/schema

resolve_goose() {
  if command -v goose >/dev/null 2>&1; then
    command -v goose
    return
  fi

  if command -v go >/dev/null 2>&1; then
    local gopath_bin
    gopath_bin="$(go env GOPATH)/bin/goose"
    if [ -x "$gopath_bin" ]; then
      echo "$gopath_bin"
      return
    fi
  fi

  echo "goose not found." >&2
  echo "Install: go install github.com/pressly/goose/v3/cmd/goose@latest" >&2
  echo "Then add \$(go env GOPATH)/bin to your PATH, or re-run this script." >&2
  exit 1
}

GOOSE_BIN="$(resolve_goose)"
exec "$GOOSE_BIN" "${@:-up}"
