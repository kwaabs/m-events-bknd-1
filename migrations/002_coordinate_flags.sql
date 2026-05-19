-- Migration 002: coordinate flags on CustomerRecords
-- ---------------------------------------------------------------------------
-- Adds two boolean columns that are kept current by database triggers:
--
--   has_own_coordinates   — true when the CustomerRecord itself has lat+lng
--   has_event_coordinates — true when any tamper event for the meter has lat+lng
--
-- A computed helper column:
--   has_any_coordinates   — true when either of the above is true
--                           (used directly by the API filter)
--
-- Triggers fire on:
--   CustomerRecords         INSERT / UPDATE  → refreshes has_own_coordinates
--   MMS_METER_TAMPER_EVENTS INSERT / UPDATE / DELETE → refreshes
--                                             has_event_coordinates on the
--                                             matching CustomerRecord row(s)
-- ---------------------------------------------------------------------------

-- ── 1. Add columns ──────────────────────────────────────────────────────────

ALTER TABLE "CustomerRecords"
    ADD COLUMN IF NOT EXISTS has_own_coordinates   BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS has_event_coordinates BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS has_any_coordinates   BOOLEAN NOT NULL DEFAULT FALSE;

-- ── 2. Back-fill existing rows ───────────────────────────────────────────────
-- Run once. May take a few minutes on 5M rows — safe to run during off-hours.

UPDATE "CustomerRecords"
SET
    has_own_coordinates = (latitude IS NOT NULL AND longitude IS NOT NULL),
    has_any_coordinates = (latitude IS NOT NULL AND longitude IS NOT NULL);

-- Back-fill has_event_coordinates in batches via a single UPDATE … FROM
UPDATE "CustomerRecords" cr
SET
    has_event_coordinates = TRUE,
    has_any_coordinates   = TRUE
    FROM (
    SELECT DISTINCT meter_number
    FROM   "MMS_METER_TAMPER_EVENTS"
    WHERE  latitude  IS NOT NULL
      AND  longitude IS NOT NULL
) ev
WHERE cr.meternumber = ev.meter_number;

-- ── 3. Indexes ───────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_cr_has_any_coordinates
    ON "CustomerRecords" (has_any_coordinates);

CREATE INDEX IF NOT EXISTS idx_cr_has_own_coordinates
    ON "CustomerRecords" (has_own_coordinates);

CREATE INDEX IF NOT EXISTS idx_cr_has_event_coordinates
    ON "CustomerRecords" (has_event_coordinates);

-- ── 4. Trigger function — fired when CustomerRecords row changes ─────────────

CREATE OR REPLACE FUNCTION trg_cr_update_own_coordinates()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.has_own_coordinates := (NEW.latitude IS NOT NULL AND NEW.longitude IS NOT NULL);
    NEW.has_any_coordinates  := NEW.has_own_coordinates OR NEW.has_event_coordinates;
RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_cr_own_coordinates ON "CustomerRecords";
CREATE TRIGGER trg_cr_own_coordinates
    BEFORE INSERT OR UPDATE OF latitude, longitude
                     ON "CustomerRecords"
                         FOR EACH ROW
                         EXECUTE FUNCTION trg_cr_update_own_coordinates();

-- ── 5. Trigger function — fired when a tamper event is inserted/updated/deleted

CREATE OR REPLACE FUNCTION trg_tamper_update_event_coordinates()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
v_meter_number TEXT;
    v_has_event    BOOLEAN;
BEGIN
    -- Determine which meter number changed
    IF TG_OP = 'DELETE' THEN
        v_meter_number := OLD.meter_number;
ELSE
        v_meter_number := NEW.meter_number;
END IF;

    -- Does this meter still have any event with coordinates?
SELECT EXISTS (
    SELECT 1
    FROM   "MMS_METER_TAMPER_EVENTS"
    WHERE  meter_number = v_meter_number
      AND  latitude     IS NOT NULL
      AND  longitude    IS NOT NULL
) INTO v_has_event;

-- Push the result to every CustomerRecord on this meter
UPDATE "CustomerRecords"
SET
    has_event_coordinates = v_has_event,
    has_any_coordinates   = (
        (latitude IS NOT NULL AND longitude IS NOT NULL) OR v_has_event
        )
WHERE meternumber = v_meter_number;

IF TG_OP = 'DELETE' THEN
        RETURN OLD;
END IF;
RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_tamper_event_coordinates ON "MMS_METER_TAMPER_EVENTS";
CREATE TRIGGER trg_tamper_event_coordinates
    AFTER INSERT OR UPDATE OF latitude, longitude OR DELETE
                    ON "MMS_METER_TAMPER_EVENTS"
                        FOR EACH ROW
                        EXECUTE FUNCTION trg_tamper_update_event_coordinates();

-- ── Done ─────────────────────────────────────────────────────────────────────
-- After this migration:
--   has_any_coordinates = TRUE  → customer can be plotted on the map
--   has_any_coordinates = FALSE → customer needs geocoding
-- Both values stay current automatically via the triggers above.