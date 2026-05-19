-- post_load.sql
-- Resumable post-load refresh. Each step commits independently.
-- If a step fails, re-running skips already-completed steps.
-- To force a full re-run: SELECT post_load_reset(); then re-run.

-- ── Checkpoint table ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS post_load_checkpoint (
                                                    step        TEXT        PRIMARY KEY,
                                                    completed   BOOLEAN     NOT NULL DEFAULT FALSE,
                                                    started_at  TIMESTAMPTZ,
                                                    finished_at TIMESTAMPTZ,
                                                    error_msg   TEXT
);

-- Ensure all steps are registered
INSERT INTO post_load_checkpoint (step, completed)
VALUES
    ('meter_summary',        FALSE),
    ('coordinate_flags',     FALSE),
    ('geom',                 FALSE),
    ('district_summary',     FALSE)
    ON CONFLICT (step) DO NOTHING;

-- Helper function to reset all checkpoints (for a forced full re-run)
CREATE OR REPLACE FUNCTION post_load_reset()
RETURNS VOID LANGUAGE SQL AS $$
UPDATE post_load_checkpoint SET completed = FALSE, finished_at = NULL, error_msg = NULL;
$$;

\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
\echo 'ECG Post-Load Refresh (resumable)'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'

-- Show current checkpoint state
SELECT
    step,
    CASE WHEN completed THEN '✓ done' ELSE '○ pending' END AS status,
    finished_at
FROM post_load_checkpoint
ORDER BY CASE step
             WHEN 'meter_summary'    THEN 1
             WHEN 'coordinate_flags' THEN 2
             WHEN 'geom'             THEN 3
             WHEN 'district_summary' THEN 4
             END;

-- ── Step 1: meter_summary ─────────────────────────────────────────────────────
DO $$
BEGIN
    IF (SELECT completed FROM post_load_checkpoint WHERE step = 'meter_summary') THEN
        RAISE NOTICE '[1/4] meter_summary already done — skipping';
ELSE
        RAISE NOTICE '[1/4] Refreshing meter_summary...';
UPDATE post_load_checkpoint
SET started_at = now(), error_msg = NULL
WHERE step = 'meter_summary';

INSERT INTO meter_summary (
    meternumber, accountnumber,
    latest_event_time, latest_event_code, latest_event_desc,
    latest_event_lat, latest_event_lng,
    event_count, has_event_coordinates, updated_at
)
SELECT DISTINCT ON (e.meter_number)
    e.meter_number,
    cr.accountnumber,
    e.event_time, e.event_code, e.event_desc,
    e.latitude, e.longitude,
    COUNT(*) OVER (PARTITION BY e.meter_number),
    (e.latitude IS NOT NULL AND e.longitude IS NOT NULL),
    now()
FROM "MMS_METER_TAMPER_EVENTS" e
    LEFT JOIN "CustomerRecords" cr ON cr.meternumber = e.meter_number
ORDER BY e.meter_number, e.event_time DESC
ON CONFLICT (meternumber) DO UPDATE SET
    accountnumber         = EXCLUDED.accountnumber,
                                 latest_event_time     = EXCLUDED.latest_event_time,
                                 latest_event_code     = EXCLUDED.latest_event_code,
                                 latest_event_desc     = EXCLUDED.latest_event_desc,
                                 latest_event_lat      = EXCLUDED.latest_event_lat,
                                 latest_event_lng      = EXCLUDED.latest_event_lng,
                                 event_count           = EXCLUDED.event_count,
                                 has_event_coordinates = EXCLUDED.has_event_coordinates,
                                 updated_at            = now();

UPDATE post_load_checkpoint
SET completed = TRUE, finished_at = now()
WHERE step = 'meter_summary';

RAISE NOTICE '[1/4] meter_summary OK';
END IF;
END $$;

-- ── Step 2: Coordinate flags ──────────────────────────────────────────────────
DO $$
BEGIN
    IF (SELECT completed FROM post_load_checkpoint WHERE step = 'coordinate_flags') THEN
        RAISE NOTICE '[2/4] coordinate_flags already done — skipping';
ELSE
        RAISE NOTICE '[2/4] Syncing coordinate flags...';
UPDATE post_load_checkpoint
SET started_at = now(), error_msg = NULL
WHERE step = 'coordinate_flags';

-- Customers that have events
UPDATE "CustomerRecords" cr
SET
    has_own_coordinates   = (cr.latitude IS NOT NULL AND cr.longitude IS NOT NULL),
    has_event_coordinates = COALESCE(ms.has_event_coordinates, FALSE),
    has_any_coordinates   = (
        (cr.latitude IS NOT NULL AND cr.longitude IS NOT NULL)
            OR COALESCE(ms.has_event_coordinates, FALSE)
        )
    FROM meter_summary ms
WHERE cr.meternumber = ms.meternumber;

-- Customers with no events
UPDATE "CustomerRecords"
SET
    has_own_coordinates   = (latitude IS NOT NULL AND longitude IS NOT NULL),
    has_event_coordinates = FALSE,
    has_any_coordinates   = (latitude IS NOT NULL AND longitude IS NOT NULL)
WHERE meternumber NOT IN (SELECT meternumber FROM meter_summary);

UPDATE post_load_checkpoint
SET completed = TRUE, finished_at = now()
WHERE step = 'coordinate_flags';

RAISE NOTICE '[2/4] coordinate_flags OK';
END IF;
END $$;

-- ── Step 3: Geometry column ───────────────────────────────────────────────────
DO $$
BEGIN
    IF (SELECT completed FROM post_load_checkpoint WHERE step = 'geom') THEN
        RAISE NOTICE '[3/4] geom already done — skipping';
ELSE
        RAISE NOTICE '[3/4] Populating geom column...';
UPDATE post_load_checkpoint
SET started_at = now(), error_msg = NULL
WHERE step = 'geom';

UPDATE "CustomerRecords"
SET geom = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
WHERE latitude IS NOT NULL AND longitude IS NOT NULL AND geom IS NULL;

UPDATE "CustomerRecords"
SET geom = NULL
WHERE (latitude IS NULL OR longitude IS NULL) AND geom IS NOT NULL;

UPDATE post_load_checkpoint
SET completed = TRUE, finished_at = now()
WHERE step = 'geom';

RAISE NOTICE '[3/4] geom OK';
END IF;
END $$;

-- ── Step 4: district_event_summary ───────────────────────────────────────────
DO $$
BEGIN
    IF (SELECT completed FROM post_load_checkpoint WHERE step = 'district_summary') THEN
        RAISE NOTICE '[4/4] district_summary already done — skipping';
ELSE
        RAISE NOTICE '[4/4] Refreshing district_event_summary...';
UPDATE post_load_checkpoint
SET started_at = now(), error_msg = NULL
WHERE step = 'district_summary';

INSERT INTO district_event_summary (
    districtcode, districtname, regioncode, regionname, geom,
    total_meters, meters_with_events, meters_with_coordinates,
    meters_without_coordinates, total_events, latest_event_time, updated_at
)
-- Pre-aggregate to one row per districtcode before the upsert.
-- Avoids "ON CONFLICT DO UPDATE command cannot affect row a second time"
-- which happens when CustomerRecords has duplicate districtcode values
-- with slightly different names/casing.
SELECT
    agg.districtcode,
    agg.districtname,
    agg.regioncode,
    agg.regionname,
    agg.geom,
    agg.total_meters,
    agg.meters_with_events,
    agg.meters_with_coordinates,
    agg.meters_without_coordinates,
    agg.total_events,
    agg.latest_event_time,
    now()
FROM (
         SELECT
             cr.districtcode,
             MAX(cr.districtname)  AS districtname,
             MAX(cr.regioncode)    AS regioncode,
             MAX(cr.regionname)    AS regionname,
             ST_Centroid(ST_Collect(cr.geom))                                       AS geom,
             COUNT(DISTINCT cr.meternumber)                                         AS total_meters,
             COUNT(DISTINCT ms.meternumber) FILTER (WHERE ms.event_count > 0)       AS meters_with_events,
             COUNT(DISTINCT cr.meternumber) FILTER (WHERE cr.has_any_coordinates)   AS meters_with_coordinates,
             COUNT(DISTINCT cr.meternumber) FILTER (WHERE NOT cr.has_any_coordinates) AS meters_without_coordinates,
             COALESCE(SUM(ms.event_count), 0)                                       AS total_events,
             MAX(ms.latest_event_time)                                              AS latest_event_time
         FROM "CustomerRecords" cr
                  LEFT JOIN meter_summary ms ON ms.meternumber = cr.meternumber
         WHERE cr.districtcode IS NOT NULL
         GROUP BY cr.districtcode
     ) agg
    ON CONFLICT (districtcode) DO UPDATE SET
    geom                       = EXCLUDED.geom,
                                      total_meters               = EXCLUDED.total_meters,
                                      meters_with_events         = EXCLUDED.meters_with_events,
                                      meters_with_coordinates    = EXCLUDED.meters_with_coordinates,
                                      meters_without_coordinates = EXCLUDED.meters_without_coordinates,
                                      total_events               = EXCLUDED.total_events,
                                      latest_event_time          = EXCLUDED.latest_event_time,
                                      updated_at                 = now();

UPDATE post_load_checkpoint
SET completed = TRUE, finished_at = now()
WHERE step = 'district_summary';

RAISE NOTICE '[4/4] district_summary OK';
END IF;
END $$;

-- ── Final summary ─────────────────────────────────────────────────────────────
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
\echo 'Checkpoint State'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
SELECT
    step,
    CASE WHEN completed THEN '✓ done' ELSE '✗ pending' END AS status,
    finished_at
FROM post_load_checkpoint
ORDER BY CASE step
             WHEN 'meter_summary'    THEN 1
             WHEN 'coordinate_flags' THEN 2
             WHEN 'geom'             THEN 3
             WHEN 'district_summary' THEN 4
             END;

\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
\echo 'Load Summary'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
SELECT
    to_char((SELECT COUNT(*) FROM "CustomerRecords"),         '999,999,999') AS total_customers,
    to_char((SELECT COUNT(*) FROM "MMS_METER_TAMPER_EVENTS"), '999,999,999') AS total_events,
    to_char((SELECT COUNT(*) FROM meter_summary),             '999,999,999') AS meters_with_events,
    to_char((SELECT COUNT(*) FROM "CustomerRecords"
             WHERE has_any_coordinates),                      '999,999,999') AS plottable_meters,
    to_char((SELECT COUNT(*) FROM "CustomerRecords"
             WHERE NOT has_any_coordinates),                  '999,999,999') AS unplottable_meters;