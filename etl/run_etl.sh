#!/usr/bin/env bash
# ETL runner — loads CSVs into the staging database.
# Usage: ./etl/run_etl.sh /path/to/customers.csv /path/to/events.csv

set -euo pipefail

CUSTOMERS_CSV="${1:?Usage: run_etl.sh <customers.csv> <events.csv>}"
EVENTS_CSV="${2:?Usage: run_etl.sh <customers.csv> <events.csv>}"
DATABASE_URL="${DATABASE_URL:?DATABASE_URL env var must be set}"

log() { echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] $*"; }

log "Starting ETL load"
log "  customers: $CUSTOMERS_CSV"
log "  events:    $EVENTS_CSV"

# Validate files exist
[[ -f "$CUSTOMERS_CSV" ]] || { log "ERROR: customers file not found: $CUSTOMERS_CSV"; exit 1; }
[[ -f "$EVENTS_CSV" ]]    || { log "ERROR: events file not found: $EVENTS_CSV"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log "Running load.sql..."
psql "$DATABASE_URL" \
    -v customers_csv="$CUSTOMERS_CSV" \
    -v events_csv="$EVENTS_CSV" \
    -f "$SCRIPT_DIR/load.sql" \
    --echo-errors \
    2>&1

log "ETL complete"