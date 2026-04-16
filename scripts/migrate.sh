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

if [ -z "${DB_URL:-}" ]; then
  echo "DB_URL is not set. Add it to .env (e.g. postgres://user:pass@localhost:5432/debtease?sslmode=disable)." >&2
  exit 1
fi

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="$DB_URL"
export GOOSE_MIGRATION_DIR=database/schema

exec goose "${@:-up}"
