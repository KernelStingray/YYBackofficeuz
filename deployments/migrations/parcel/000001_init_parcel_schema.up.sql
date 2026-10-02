CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS parcels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID NOT NULL,
    driver_id UUID, -- NULL until accepted
    origin GEOMETRY(Point, 4326) NOT NULL,
    destination GEOMETRY(Point, 4326) NOT NULL,
    receiver_phone_encrypted VARCHAR(255) NOT NULL,
    receiver_name VARCHAR(100),
    dropoff_type VARCHAR(20) DEFAULT 'highway',
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'accepted', 'picked_up', 'delivered', 'canceled'
    price NUMERIC NOT NULL,
    origin_address VARCHAR(500),
    destination_address VARCHAR(500),
    description TEXT,
    size VARCHAR(50),
    deliver_by TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Index for searching nearby pending parcels (Spatial index)
CREATE INDEX idx_parcels_status ON parcels(status);
CREATE INDEX idx_parcels_origin ON parcels USING GIST (origin);
CREATE INDEX idx_parcels_destination ON parcels USING GIST (destination);

CREATE TABLE parcel_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parcel_id UUID NOT NULL REFERENCES parcels(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    offered_price NUMERIC,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
