-- migrations/001_create_tables.sql
-- Run this once against your Postgres database.
-- If the tables already exist (migrated from MongoDB), skip the CREATE TABLE
-- blocks and only apply the indexes at the bottom.

-- ─── CustomerRecords ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS "CustomerRecords" (
                                                 ecgkey                  TEXT PRIMARY KEY,
                                                 _id                     TEXT,
                                                 regioncode              TEXT,
                                                 regionname              TEXT,
                                                 districtcode            TEXT,
                                                 districtname            TEXT,
                                                 servicetype             TEXT,
                                                 serviceclass            TEXT,
                                                 tariffclasscode         TEXT,
                                                 tariffclassname         TEXT,
                                                 community               TEXT,
                                                 contracteddemand        NUMERIC,
                                                 qrcode                  TEXT,
                                                 createdat               TIMESTAMPTZ,
                                                 updatedat               TIMESTAMPTZ,
                                                 address                 TEXT,
                                                 ghanapostaddress        TEXT,
                                                 streetname              TEXT,
                                                 housenumber             TEXT,
                                                 geolocationtype         TEXT,
                                                 geolocationid           TEXT,
                                                 latitude                DOUBLE PRECISION,
                                                 longitude               DOUBLE PRECISION,
                                                 emailaddress            TEXT,
                                                 phonenumber             TEXT,
                                                 profileimageurl         TEXT,
                                                 fullname                TEXT,
                                                 ghanacardnumber         TEXT,
                                                 payerfullname           TEXT,
                                                 payerprimaryphonenumber TEXT,
                                                 payerfirstphone         TEXT,
                                                 payersecondphone        TEXT,
                                                 servicepointnumber      TEXT,
                                                 accountnumber           TEXT,
                                                 contractstatus          TEXT,
                                                 ismigrateddata          BOOLEAN DEFAULT FALSE,
                                                 metertype               TEXT,
                                                 meterlocation           TEXT,
                                                 activity                TEXT,
                                                 subactivity             TEXT,
                                                 customertype            TEXT,
                                                 isregularized           BOOLEAN DEFAULT FALSE,
                                                 lastreadingvalue        BIGINT,
                                                 geolocationcoordinates  TEXT,
                                                 payerphonenumbers       TEXT,
                                                 blockcode               TEXT,
                                                 blockname               TEXT,
                                                 roundcode               TEXT,
                                                 roundname               TEXT,
                                                 cmscontractstatus       TEXT,
                                                 code                    TEXT,
                                                 geocode                 TEXT,
                                                 metermake               TEXT,
                                                 metermodel              TEXT,
                                                 meternumber             TEXT,
                                                 meterphase              TEXT,
                                                 propertycode            TEXT,
                                                 plotcode                TEXT,
                                                 ministry                TEXT,
                                                 mda                     TEXT,
                                                 lastreadingdate         TIMESTAMPTZ,
                                                 installationdate        TIMESTAMPTZ,
                                                 lastbillamount          NUMERIC,
                                                 lastbillconsumption     NUMERIC,
                                                 lastpaymentdate         DATE,
                                                 lastpaymentamount       NUMERIC,
                                                 currentbalance          NUMERIC,
                                                 nationality             TEXT,
                                                 propertyownerfullname   TEXT,
                                                 propertyownerphonenumber TEXT,
                                                 accounttype             TEXT,
                                                 isamr                   BOOLEAN DEFAULT FALSE,
                                                 "offset"                TEXT,
                                                 idnumber                TEXT,
                                                 ministrycode            TEXT,
                                                 ministryname            TEXT,
                                                 mdacode                 TEXT,
                                                 mdaname                 TEXT,
                                                 extraattributesname     TEXT,
                                                 extraattributesvalue    BOOLEAN DEFAULT FALSE,
                                                 extraattributesid       TEXT,
                                                 lastbilldate            TIMESTAMPTZ,
                                                 geolocation             JSONB
);

-- ─── MMS_METER_TAMPER_EVENTS ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS "MMS_METER_TAMPER_EVENTS" (
                                                         period        TEXT,
                                                         meter_number  TEXT        NOT NULL,
                                                         customer_name TEXT,
                                                         event_code    TEXT,
                                                         event_desc    TEXT,
                                                         event_time    TIMESTAMPTZ NOT NULL,
                                                         latitude      DOUBLE PRECISION,
                                                         longitude     DOUBLE PRECISION,
                                                         counting      BIGINT
);

-- ─── Indexes (critical for join & search performance) ────────────────────────

-- CustomerRecords
CREATE INDEX IF NOT EXISTS idx_cr_meternumber
    ON "CustomerRecords" (meternumber);

CREATE INDEX IF NOT EXISTS idx_cr_accountnumber
    ON "CustomerRecords" (accountnumber);

CREATE INDEX IF NOT EXISTS idx_cr_fullname
    ON "CustomerRecords" (fullname);

CREATE INDEX IF NOT EXISTS idx_cr_regioncode
    ON "CustomerRecords" (regioncode);

CREATE INDEX IF NOT EXISTS idx_cr_districtcode
    ON "CustomerRecords" (districtcode);

CREATE INDEX IF NOT EXISTS idx_cr_contractstatus
    ON "CustomerRecords" (contractstatus);

CREATE INDEX IF NOT EXISTS idx_cr_servicetype
    ON "CustomerRecords" (servicetype);

-- For free-text search using ILIKE / UPPER() LIKE
CREATE INDEX IF NOT EXISTS idx_cr_fullname_upper
    ON "CustomerRecords" (UPPER(fullname));

-- MMS_METER_TAMPER_EVENTS
CREATE INDEX IF NOT EXISTS idx_tamper_meter_number
    ON "MMS_METER_TAMPER_EVENTS" (meter_number);

CREATE INDEX IF NOT EXISTS idx_tamper_event_time
    ON "MMS_METER_TAMPER_EVENTS" (event_time DESC);

-- Composite index for the lateral join: meter_number + event_time DESC
CREATE INDEX IF NOT EXISTS idx_tamper_meter_event_time
    ON "MMS_METER_TAMPER_EVENTS" (meter_number, event_time DESC);

-- ─── Coordinate indexes ───────────────────────────────────────────────────────

-- Partial index: customers WITH coordinates (for has_coordinates filter)
CREATE INDEX IF NOT EXISTS idx_cr_has_coordinates
    ON "CustomerRecords" (meternumber)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

-- Partial index: customers WITHOUT coordinates (for no_coordinates filter)
CREATE INDEX IF NOT EXISTS idx_cr_no_coordinates
    ON "CustomerRecords" (meternumber)
    WHERE latitude IS NULL OR longitude IS NULL;

-- Events with coordinates (used by lateral coordinate checks)
CREATE INDEX IF NOT EXISTS idx_tamper_has_coordinates
    ON "MMS_METER_TAMPER_EVENTS" (meter_number)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;