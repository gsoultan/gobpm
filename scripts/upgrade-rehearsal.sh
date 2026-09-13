#!/usr/bin/env bash
#
# Rehearse an upgrade against a copy of real data, before doing it for real.
#
# The test suite starts from an empty schema. That proves a fresh install works
# and says nothing about the case every existing installation is in — a database
# with two years of rows, whose upgrade path runs migrations that a fresh
# install skips entirely. tests/upgrade automates the shape of that; this runs
# it against *your* data, which is the part no test can do for you.
#
# It is read-only with respect to production: everything happens in a scratch
# database created from a backup, and the scratch database is dropped at the end
# unless you ask to keep it.
#
# Usage:
#   scripts/upgrade-rehearsal.sh <backup-directory> [--keep]
#
# Reads from the environment:
#   PGHOST PGPORT PGUSER PGPASSWORD   where to create the scratch database
#   ENCRYPTION_KEY                    the key the backup's data was encrypted with
#   METIS_REHEARSAL_DB                scratch database name (default metis_rehearsal)
#
# Exits non-zero on the first thing that would have gone wrong in production.
set -euo pipefail

SOURCE="${1:-}"
KEEP="${2:-}"
SCRATCH="${METIS_REHEARSAL_DB:-metis_rehearsal}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${METIS_REHEARSAL_PORT:-18080}"

die()  { printf '\033[31mrehearsal: %s\033[0m\n' "$1" >&2; exit 1; }
note() { printf '\033[36mrehearsal:\033[0m %s\n' "$1"; }
pass() { printf '\033[32m  ok\033[0m   %s\n' "$1"; }

[ -n "$SOURCE" ] || die "usage: scripts/upgrade-rehearsal.sh <backup-directory> [--keep]"
[ -d "$SOURCE" ] || die "no such backup directory: $SOURCE"
[ -n "${ENCRYPTION_KEY:-}" ] || die "ENCRYPTION_KEY is required: the backup's rows are encrypted, and restoring them without it proves nothing"
command -v psql >/dev/null || die "psql not found"

DSN="host=${PGHOST:-127.0.0.1} port=${PGPORT:-5432} user=${PGUSER:-postgres} password=${PGPASSWORD:-} dbname=${SCRATCH} sslmode=${PGSSLMODE:-disable}"
URL="postgres://${PGUSER:-postgres}:${PGPASSWORD:-}@${PGHOST:-127.0.0.1}:${PGPORT:-5432}/${SCRATCH}?sslmode=${PGSSLMODE:-disable}"

cleanup() {
  status=$?
  if [ -n "${SERVER_PID:-}" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  if [ "$KEEP" != "--keep" ]; then
    psql -d postgres -qc "DROP DATABASE IF EXISTS ${SCRATCH}" >/dev/null 2>&1 || true
  else
    note "kept ${SCRATCH} for inspection"
  fi
  exit $status
}
trap cleanup EXIT

# 1 — a scratch copy. Never the real database: a rehearsal that can damage what
# it is rehearsing for is not a rehearsal.
note "creating ${SCRATCH} from ${SOURCE}"
psql -d postgres -qc "DROP DATABASE IF EXISTS ${SCRATCH}" >/dev/null
psql -d postgres -qc "CREATE DATABASE ${SCRATCH}" >/dev/null
PGDATABASE="$SCRATCH" "${ROOT}/scripts/restore.sh" "$SOURCE" >/dev/null \
  || die "the backup would not restore — fix that before upgrading anything"
pass "restored"

BEFORE="$(psql -d "$SCRATCH" -tAc "SELECT count(*) FROM process_instances" 2>/dev/null || echo 0)"
note "the copy holds ${BEFORE} process instances"

# 2 — the build under test, booted against the copy. This is the migration run,
# the storm table creation and the column-default reconciliation, in the order
# the real boot performs them.
note "booting the build under test"
( cd "$ROOT" && go build -o /tmp/metis-rehearsal ./cmd/metis ) || die "the build failed"

DATABASE_URL="$URL" ENCRYPTION_KEY="$ENCRYPTION_KEY" \
  METIS_HTTP_ADDRESS=":${PORT}" JWT_SECRET="${JWT_SECRET:-rehearsal-only-secret-not-for-production-use}" \
  /tmp/metis-rehearsal > /tmp/metis-rehearsal.log 2>&1 &
SERVER_PID=$!

for _ in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:${PORT}/readyz" >/dev/null 2>&1; then break; fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    tail -40 /tmp/metis-rehearsal.log >&2
    die "the server exited during startup — this is the upgrade failing, and it would have failed in production"
  fi
  sleep 1
done
curl -fsS "http://127.0.0.1:${PORT}/readyz" >/dev/null 2>&1 \
  || { tail -40 /tmp/metis-rehearsal.log >&2; die "the server never became ready"; }
pass "migrated and ready"

# 3 — the checks a green readiness probe does not cover.
#
# readyz answers "the process is up", which an upgrade that silently emptied a
# column also satisfies. Each of these is a thing that has actually broken.

if grep -q "does not match the model" /tmp/metis-rehearsal.log; then
  note "schema drift reported at startup:"
  grep "does not match the model" /tmp/metis-rehearsal.log | sed 's/^/    /'
  die "the upgraded schema disagrees with the model — reconcile before upgrading production"
fi
pass "no schema drift"

AFTER="$(psql -d "$SCRATCH" -tAc "SELECT count(*) FROM process_instances")"
[ "$AFTER" = "$BEFORE" ] || die "process instances went from ${BEFORE} to ${AFTER} across the upgrade"
pass "every process instance survived (${AFTER})"

# The columns migration 18 moves. Empty here means the data was left in a column
# the new code does not read — which is silent, and is the bug this exists for.
for check in \
  "forms:fields" \
  "connectors:properties" \
  "external_tasks:instance_id" \
  "user_organizations:user_id"
do
  table="${check%%:*}"; column="${check##*:}"
  total="$(psql -d "$SCRATCH" -tAc "SELECT count(*) FROM ${table}" 2>/dev/null || echo 0)"
  [ "$total" -gt 0 ] || continue
  filled="$(psql -d "$SCRATCH" -tAc "SELECT count(*) FROM ${table} WHERE ${column} IS NOT NULL")"
  [ "$filled" = "$total" ] \
    || die "${table}.${column} is empty on $((total - filled)) of ${total} rows — the upgrade moved the column and not the data"
  pass "${table}.${column} carried its data (${filled}/${total})"
done

# Somebody can still sign in and see their own work. An upgrade that leaves
# every account without its organization passes every check above.
ORPHANS="$(psql -d "$SCRATCH" -tAc \
  "SELECT count(*) FROM users u WHERE u.deleted_at IS NULL
     AND NOT EXISTS (SELECT 1 FROM user_organizations o WHERE o.user_id = u.id)")"
[ "$ORPHANS" = "0" ] || die "${ORPHANS} accounts have no organization after the upgrade — they would sign in and see nothing"
pass "every account kept its organization"

printf '\n\033[32mrehearsal passed\033[0m — this build upgrades this data.\n'
printf 'What it does not cover: load, your ingress, and anything only your users do.\n'
