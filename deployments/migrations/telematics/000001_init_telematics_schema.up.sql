CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

CREATE TABLE driver_location_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL,
    geom GEOMETRY(POINT, 4326) NOT NULL,
    speed NUMERIC DEFAULT 0,
    heading NUMERIC DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Index for searching driver history
CREATE INDEX idx_driver_location_history_driver_id ON driver_location_history(driver_id);
-- Spatial index for analyzing locations
CREATE INDEX idx_driver_location_history_geom ON driver_location_history USING GIST (geom);
-- Index for time series queries
CREATE INDEX idx_driver_location_history_created_at ON driver_location_history(created_at);
