-- Migration 003: TimescaleDB hypertable for events
-- Supabase Postgres ships the Apache-licensed TimescaleDB.
-- Compression is a community/paid feature — omitted here.

-- ── Convert tamper events to TimescaleDB hypertable ──────────────────────────
-- event_time must be NOT NULL for TimescaleDB
ALTER TABLE "MMS_METER_TAMPER_EVENTS"
    ALTER COLUMN event_time SET NOT NULL;

-- Convert to hypertable partitioned monthly on event_time.
-- This gives Postgres partition pruning on date-range queries — the most
-- common query pattern on 29M+ events.
SELECT create_hypertable(
               '"MMS_METER_TAMPER_EVENTS"',
               'event_time',
               chunk_time_interval => INTERVAL '1 month',
               if_not_exists       => TRUE
       );

-- ── Indexes ───────────────────────────────────────────────────────────────────

-- Most critical: powers the lateral join (latest event per meter)
CREATE INDEX IF NOT EXISTS idx_tamper_meter_event_time
    ON "MMS_METER_TAMPER_EVENTS" (meter_number, event_time DESC);

-- Event description search
CREATE INDEX IF NOT EXISTS idx_tamper_event_desc
    ON "MMS_METER_TAMPER_EVENTS" (event_time DESC)
    WHERE event_desc IS NOT NULL;

-- Coordinate presence filter
CREATE INDEX IF NOT EXISTS idx_tamper_has_coordinates
    ON "MMS_METER_TAMPER_EVENTS" (meter_number)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;