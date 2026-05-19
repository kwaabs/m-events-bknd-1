-- Migration 000: extensions + clean slate
-- Safe to run multiple times.

CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Verify
SELECT name, installed_version
FROM pg_available_extensions
WHERE name IN ('postgis', 'timescaledb', 'postgis_topology')
  AND installed_version IS NOT NULL;

-- Create a read-only user for Martin tile server
-- This user can connect via TCP from any Docker container
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'martin_reader') THEN
CREATE ROLE martin_reader WITH LOGIN PASSWORD 'martin_tiles_readonly';
END IF;
END $$;

-- Grant read access to the tables Martin needs
GRANT CONNECT ON DATABASE ecg TO martin_reader;
GRANT USAGE ON SCHEMA public TO martin_reader;
GRANT SELECT ON "CustomerRecords" TO martin_reader;
GRANT SELECT ON district_event_summary TO martin_reader;
GRANT SELECT ON meter_summary TO martin_reader;
GRANT SELECT ON customer_map_view TO martin_reader;
GRANT SELECT ON account_meters TO martin_reader;