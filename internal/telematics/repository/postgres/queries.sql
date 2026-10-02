-- name: BulkInsertLocations :exec
INSERT INTO driver_location_history (driver_id, geom, speed, heading, created_at)
SELECT 
    unnest(@driver_ids::uuid[]),
    ST_SetSRID(ST_MakePoint(unnest(@longitudes::float8[]), unnest(@latitudes::float8[])), 4326),
    unnest(@speeds::numeric[]),
    unnest(@headings::numeric[]),
    unnest(@created_ats::timestamptz[])
;

-- name: GetDriverLocationHistory :many
SELECT 
    id, 
    driver_id, 
    ST_X(geom::geometry) AS longitude, 
    ST_Y(geom::geometry) AS latitude, 
    speed, 
    heading, 
    created_at
FROM driver_location_history
WHERE driver_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 100;
