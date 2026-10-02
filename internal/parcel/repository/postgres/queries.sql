-- name: CreateParcel :one
INSERT INTO parcels (
    id, sender_id, origin, destination, receiver_phone_encrypted, price, status,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type
) VALUES (
    $1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), ST_SetSRID(ST_MakePoint($5, $6), 4326), $7, $8, 'pending',
    $9, $10, $11, $12, $13, $14, $15
) RETURNING id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type;

-- name: GetParcelByID :one
SELECT 
    id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type
FROM parcels WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateParcelStatus :one
UPDATE parcels
SET status = $1, driver_id = COALESCE($2, driver_id), updated_at = NOW()
WHERE id = $3 AND deleted_at IS NULL
RETURNING id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type;

-- name: SearchParcels :many
SELECT 
    id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type
FROM parcels
WHERE status = 'pending' 
  AND ST_DWithin(
      origin::geography, 
      ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, 
      $3
  )
  AND (sqlc.narg('price_min')::numeric IS NULL OR price >= sqlc.narg('price_min')::numeric)
  AND (sqlc.narg('price_max')::numeric IS NULL OR price <= sqlc.narg('price_max')::numeric)
  AND deleted_at IS NULL;

-- name: SearchParcelsInCorridor :many
SELECT 
    id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type
FROM parcels
WHERE status = 'pending' 
  AND ST_DWithin(
      origin::geography, 
      ST_MakeLine(ST_SetSRID(ST_MakePoint($1, $2), 4326), ST_SetSRID(ST_MakePoint($3, $4), 4326))::geography, 
      $5
  )
  AND ST_DWithin(
      destination::geography, 
      ST_MakeLine(ST_SetSRID(ST_MakePoint($1, $2), 4326), ST_SetSRID(ST_MakePoint($3, $4), 4326))::geography, 
      $5
  )
  AND ST_LineLocatePoint(
      ST_MakeLine(ST_SetSRID(ST_MakePoint($1, $2), 4326), ST_SetSRID(ST_MakePoint($3, $4), 4326)),
      origin::geometry
  ) <= ST_LineLocatePoint(
      ST_MakeLine(ST_SetSRID(ST_MakePoint($1, $2), 4326), ST_SetSRID(ST_MakePoint($3, $4), 4326)),
      destination::geometry
  )
  AND (sqlc.narg('price_min')::numeric IS NULL OR price >= sqlc.narg('price_min')::numeric)
  AND (sqlc.narg('price_max')::numeric IS NULL OR price <= sqlc.narg('price_max')::numeric)
  AND deleted_at IS NULL;

-- name: DeleteParcel :exec
UPDATE parcels SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAllParcels :many
SELECT 
    id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type
FROM parcels
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetParcelsBySenderID :many
SELECT 
    id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type,
    (SELECT COUNT(*)::int FROM parcel_responses pr WHERE pr.parcel_id = parcels.id) AS responses_count
FROM parcels
WHERE sender_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetParcelsByDriverID :many
SELECT 
    id, sender_id, driver_id, receiver_phone_encrypted, status, price, created_at, updated_at,
    ST_X(origin::geometry) AS origin_lon,
    ST_Y(origin::geometry) AS origin_lat,
    ST_X(destination::geometry) AS destination_lon,
    ST_Y(destination::geometry) AS destination_lat,
    origin_address, destination_address, description, size, deliver_by, receiver_name, dropoff_type
FROM parcels
WHERE driver_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateParcelResponse :one
INSERT INTO parcel_responses (
    id, parcel_id, driver_id, status, offered_price
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, parcel_id, driver_id, status, offered_price, created_at, updated_at;

-- name: GetParcelResponsesByParcelID :many
SELECT 
    id, parcel_id, driver_id, status, offered_price, created_at, updated_at
FROM parcel_responses
WHERE parcel_id = $1
ORDER BY created_at ASC;

-- name: GetTotalParcelsCount :one
SELECT count(*) FROM parcels WHERE deleted_at IS NULL;

-- name: GetActiveParcelsCount :one
SELECT count(*) FROM parcels WHERE status IN ('pending', 'accepted') AND deleted_at IS NULL;

-- name: GetDeliveredParcelsCount :one
SELECT count(*) FROM parcels WHERE status = 'delivered' AND deleted_at IS NULL;

-- name: GetParcelsChartData :many
SELECT DATE(created_at)::text as date, count(*)::bigint as count 
FROM parcels 
WHERE created_at >= NOW() - INTERVAL '7 days' AND deleted_at IS NULL
GROUP BY DATE(created_at) 
ORDER BY date ASC;

-- name: CountUserParcelsAsSender :one
SELECT COUNT(*) FROM parcels
WHERE sender_id = $1 AND deleted_at IS NULL;

-- name: CountUserParcelsAsDriver :one
SELECT COUNT(*) FROM parcels
WHERE driver_id = $1 AND deleted_at IS NULL;
