-- Migration 004: PostGIS geometry column, meter_summary, views for Martin
-- Depends on 003 (PostGIS extension must exist)

-- ── 1. Add geometry column to CustomerRecords ─────────────────────────────────

ALTER TABLE "CustomerRecords"
    ADD COLUMN IF NOT EXISTS geom geometry(Point, 4326);

-- Back-fill from existing lat/lng
-- Source data has lat/lng columns swapped — correct at geom level
UPDATE "CustomerRecords"
SET geom = ST_SetSRID(ST_MakePoint(latitude, longitude), 4326)
WHERE latitude  IS NOT NULL
  AND longitude IS NOT NULL
  AND geom      IS NULL;

-- Spatial index — enables viewport (bounding box) queries in Martin and API
CREATE INDEX IF NOT EXISTS idx_cr_geom
    ON "CustomerRecords" USING GIST (geom);

-- Trigger: keep geom current when ETL updates lat/lng
CREATE OR REPLACE FUNCTION trg_cr_update_geom()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.latitude IS NOT NULL AND NEW.longitude IS NOT NULL THEN
        -- Source data has lat/lng columns swapped — correct at geom level
        NEW.geom                := ST_SetSRID(ST_MakePoint(NEW.latitude, NEW.longitude), 4326);
        NEW.has_own_coordinates := TRUE;
ELSE
        NEW.geom                := NULL;
        NEW.has_own_coordinates := FALSE;
END IF;
    NEW.has_any_coordinates := NEW.has_own_coordinates OR COALESCE(NEW.has_event_coordinates, FALSE);
RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_cr_geom ON "CustomerRecords";
CREATE TRIGGER trg_cr_geom
    BEFORE INSERT OR UPDATE OF latitude, longitude
                     ON "CustomerRecords"
                         FOR EACH ROW EXECUTE FUNCTION trg_cr_update_geom();

-- ── 2. meter_summary table ────────────────────────────────────────────────────
-- One row per meter. Updated by ETL after each data load.
-- Eliminates the lateral join on the main customer query.

CREATE TABLE IF NOT EXISTS meter_summary (
                                             meternumber           TEXT        PRIMARY KEY,
                                             accountnumber         TEXT,
                                             latest_event_time     TIMESTAMPTZ,
                                             latest_event_code     TEXT,
                                             latest_event_desc     TEXT,
                                             latest_event_lat      DOUBLE PRECISION,
                                             latest_event_lng      DOUBLE PRECISION,
                                             event_count           BIGINT      NOT NULL DEFAULT 0,
                                             has_event_coordinates BOOLEAN     NOT NULL DEFAULT FALSE,
                                             updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_ms_accountnumber
    ON meter_summary (accountnumber);

CREATE INDEX IF NOT EXISTS idx_ms_has_event_coordinates
    ON meter_summary (has_event_coordinates);

CREATE INDEX IF NOT EXISTS idx_ms_latest_event_time
    ON meter_summary (latest_event_time DESC);

-- ── 3. District summary table ─────────────────────────────────────────────────
-- Pre-aggregated per district. Refreshed by ETL. Powers dashboard cards
-- and Martin's low-zoom choropleth layer.

CREATE TABLE IF NOT EXISTS district_event_summary (
                                                      districtcode                TEXT        PRIMARY KEY,
                                                      districtname                TEXT,
                                                      regioncode                  TEXT,
                                                      regionname                  TEXT,
    -- centroid of all plotable meters in the district (for Martin tile source)
                                                      geom                        geometry(Point, 4326),
    total_meters                BIGINT      NOT NULL DEFAULT 0,
    meters_with_events          BIGINT      NOT NULL DEFAULT 0,
    meters_with_coordinates     BIGINT      NOT NULL DEFAULT 0,
    meters_without_coordinates  BIGINT      NOT NULL DEFAULT 0,
    total_events                BIGINT      NOT NULL DEFAULT 0,
    prepaid_meters              BIGINT      NOT NULL DEFAULT 0,
    postpaid_meters             BIGINT      NOT NULL DEFAULT 0,
    active_contracts            BIGINT      NOT NULL DEFAULT 0,
    latest_event_time           TIMESTAMPTZ,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_des_regioncode
    ON district_event_summary (regioncode);

CREATE INDEX IF NOT EXISTS idx_des_geom
    ON district_event_summary USING GIST (geom);

-- ── 4. customer_map_view — what Martin reads for the main tile layer ──────────

CREATE OR REPLACE VIEW customer_map_view AS
SELECT
    cr.ecgkey,
    cr.meternumber,
    cr.accountnumber,
    cr.fullname,
    cr.servicetype,
    cr.serviceclass,
    cr.contractstatus,
    cr.cmscontractstatus,
    cr.regioncode,
    cr.regionname,
    cr.districtcode,
    cr.districtname,
    cr.metermake,
    cr.metermodel,
    cr.metertype,
    cr.has_any_coordinates,
    -- Best available geometry: own coords first, fall back to latest event coords
    -- ST_SetSRID on the outer COALESCE ensures SRID=4326 is always set,
    -- which Martin requires to serve the view as a tile source.
    -- Event lat/lng are also swapped in source data — corrected here.
    ST_SetSRID(
            COALESCE(
                    cr.geom,
                    CASE
                        WHEN ms.latest_event_lat IS NOT NULL AND ms.latest_event_lng IS NOT NULL
                            THEN ST_MakePoint(ms.latest_event_lng, ms.latest_event_lat)
                        END
            ),
            4326
    ) AS geom,
    ms.latest_event_time,
    ms.latest_event_code,
    ms.latest_event_desc,
    ms.latest_event_lat,
    ms.latest_event_lng,
    COALESCE(ms.event_count, 0) AS event_count
FROM "CustomerRecords" cr
         LEFT JOIN meter_summary ms ON ms.meternumber = cr.meternumber
-- Only include meters that can be plotted on the map
WHERE cr.geom IS NOT NULL
   OR (ms.latest_event_lat IS NOT NULL AND ms.latest_event_lng IS NOT NULL);

-- ── 5. account_meters view — all meters for an account ───────────────────────

CREATE OR REPLACE VIEW account_meters AS
SELECT
    cr.accountnumber,
    cr.fullname            AS customer_name,
    cr.phonenumber         AS phone,
    cr.regionname,
    cr.districtname,
    COUNT(*)               AS meter_count,
    array_agg(cr.meternumber ORDER BY cr.meternumber) AS meters,
    SUM(COALESCE(ms.event_count, 0))                  AS total_events,
    MAX(ms.latest_event_time)                         AS latest_event_time,
    bool_or(cr.has_any_coordinates)                   AS any_meter_has_coordinates
FROM "CustomerRecords" cr
         LEFT JOIN meter_summary ms ON ms.meternumber = cr.meternumber
GROUP BY
    cr.accountnumber,
    cr.fullname,
    cr.phonenumber,
    cr.regionname,
    cr.districtname;