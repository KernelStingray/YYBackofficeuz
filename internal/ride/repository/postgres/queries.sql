-- name: CreateRide :one
INSERT INTO rides (
    id, driver_id, vehicle_id, vehicle_data, origin, destination, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, has_ac, has_luggage, takes_parcels, pickup_point, pickup_lat, pickup_lon, music_taste, no_smoking, no_pets, driving_speed, max_two_back, status, is_negotiable, has_fair_price_badge, allows_city_detour, city_detour_fee
) VALUES (
    $1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326), ST_SetSRID(ST_MakePoint($7, $8), 4326), $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31
) RETURNING id, driver_id, vehicle_id, vehicle_data, ST_X(origin::geometry) as origin_lon, ST_Y(origin::geometry) as origin_lat, ST_X(destination::geometry) as dest_lon, ST_Y(destination::geometry) as dest_lat, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, has_ac, has_luggage, takes_parcels, pickup_point, pickup_lat, pickup_lon, music_taste, no_smoking, no_pets, driving_speed, max_two_back, status, is_negotiable, has_fair_price_badge, allows_city_detour, city_detour_fee, cancel_reason, cancelled_at, created_at, updated_at;

-- name: SearchRides :many
WITH ride_points AS (
    SELECT id as ride_id, origin as location, 0 as order_index
    FROM rides WHERE rides.status = 'active' AND rides.deleted_at IS NULL AND rides.available_seats >= $1
      AND rides.departure_time >= NOW() - INTERVAL '3 hours'
    UNION ALL
    SELECT id as ride_id, destination as location, 9999 as order_index
    FROM rides WHERE rides.status = 'active' AND rides.deleted_at IS NULL AND rides.available_seats >= $1
      AND rides.departure_time >= NOW() - INTERVAL '3 hours'
    UNION ALL
    SELECT rs.ride_id, rs.location, rs.order_index + 1
    FROM ride_stops rs
    INNER JOIN rides r ON r.id = rs.ride_id
    WHERE r.status = 'active' AND r.deleted_at IS NULL AND rs.deleted_at IS NULL AND r.available_seats >= $1
      AND r.departure_time >= NOW() - INTERVAL '3 hours'
)
SELECT DISTINCT 
    r.id, r.driver_id, r.vehicle_id, r.vehicle_data, ST_X(r.origin::geometry) as origin_lon, ST_Y(r.origin::geometry) as origin_lat, ST_X(r.destination::geometry) as dest_lon, ST_Y(r.destination::geometry) as dest_lat, r.origin_address, r.destination_address, r.departure_time, r.total_seats, r.available_seats, r.price_per_seat, r.gender_restriction, r.has_ac, r.has_luggage, r.takes_parcels, r.pickup_point, r.pickup_lat, r.pickup_lon, r.music_taste, r.no_smoking, r.no_pets, r.driving_speed, r.max_two_back, r.status, r.is_negotiable, r.has_fair_price_badge, r.allows_city_detour, r.city_detour_fee, r.cancel_reason, r.cancelled_at, r.created_at, r.updated_at
FROM rides r
INNER JOIN ride_points p1 ON r.id = p1.ride_id
INNER JOIN ride_points p2 ON r.id = p2.ride_id
WHERE p1.order_index < p2.order_index
  AND (
      (ST_DWithin(p1.location::geography, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4)
       AND ST_DWithin(p2.location::geography, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography, $4))
      OR
      (sqlc.narg('pickup_preference')::text = 'wait_at_route' 
       AND ST_Intersects(r.route_polygon, ST_SetSRID(ST_MakePoint($2, $3), 4326))
       AND ST_Intersects(r.route_polygon, ST_SetSRID(ST_MakePoint($5, $6), 4326)))
  )
  AND r.deleted_at IS NULL
  AND r.departure_time >= NOW() - INTERVAL '3 hours'
  AND (sqlc.narg('include_test')::boolean = true OR r.driver_id != '00000000-0000-0000-0000-000000000001'::uuid)
  AND (
      r.gender_restriction IS NULL 
      OR r.gender_restriction = ''
      OR r.gender_restriction = 'any'
      OR r.gender_restriction = sqlc.narg('passenger_gender')::text
  )
  AND (sqlc.narg('price_min')::numeric IS NULL OR r.price_per_seat >= sqlc.narg('price_min')::numeric)
  AND (sqlc.narg('price_max')::numeric IS NULL OR r.price_per_seat <= sqlc.narg('price_max')::numeric)
  AND (sqlc.narg('date_from')::timestamptz IS NULL OR r.departure_time >= sqlc.narg('date_from')::timestamptz)
  AND (sqlc.narg('date_to')::timestamptz IS NULL OR r.departure_time <= sqlc.narg('date_to')::timestamptz)
ORDER BY r.has_fair_price_badge DESC, r.departure_time ASC
LIMIT $7 OFFSET $8;

-- name: GetRideByID :one
SELECT 
    id, driver_id, vehicle_id, vehicle_data, ST_X(origin::geometry) as origin_lon, ST_Y(origin::geometry) as origin_lat, ST_X(destination::geometry) as dest_lon, ST_Y(destination::geometry) as dest_lat, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, has_ac, has_luggage, takes_parcels, pickup_point, pickup_lat, pickup_lon, music_taste, no_smoking, no_pets, driving_speed, max_two_back, status, is_negotiable, has_fair_price_badge, allows_city_detour, city_detour_fee, cancel_reason, cancelled_at, created_at, updated_at
FROM rides
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateRideStatus :exec
UPDATE rides
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: InsertRideStop :exec
INSERT INTO ride_stops (id, ride_id, location, address, order_index)
VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6);

-- name: GetRideForUpdate :one
SELECT 
    id, driver_id, vehicle_id, vehicle_data, ST_X(origin::geometry) as origin_lon, ST_Y(origin::geometry) as origin_lat, ST_X(destination::geometry) as dest_lon, ST_Y(destination::geometry) as dest_lat, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, has_ac, has_luggage, takes_parcels, pickup_point, pickup_lat, pickup_lon, music_taste, no_smoking, no_pets, driving_speed, max_two_back, status, is_negotiable, has_fair_price_badge, allows_city_detour, city_detour_fee, cancel_reason, cancelled_at, created_at, updated_at
FROM rides
WHERE id = $1 AND status = 'active' AND deleted_at IS NULL
FOR UPDATE;

-- name: GetUserMostVisitedDestinations :many
SELECT origin_address, destination_address, COUNT(*) as visit_count
FROM bookings b
JOIN rides r ON b.ride_id = r.id
WHERE b.passenger_id = $1 AND b.status = 'confirmed'
GROUP BY origin_address, destination_address
ORDER BY visit_count DESC
LIMIT $2;

-- name: GetTopGlobalDestinations :many
SELECT origin_address, destination_address, COUNT(*) as visit_count
FROM bookings b
JOIN rides r ON b.ride_id = r.id
WHERE b.status = 'confirmed'
GROUP BY origin_address, destination_address
ORDER BY visit_count DESC
LIMIT $1;

-- name: GetActiveRidesToDestinations :many
SELECT 
    id, driver_id, vehicle_id, vehicle_data, ST_X(origin::geometry) as origin_lon, ST_Y(origin::geometry) as origin_lat, ST_X(destination::geometry) as dest_lon, ST_Y(destination::geometry) as dest_lat, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, has_ac, has_luggage, takes_parcels, pickup_point, pickup_lat, pickup_lon, music_taste, no_smoking, no_pets, driving_speed, max_two_back, status, is_negotiable, has_fair_price_badge, allows_city_detour, city_detour_fee, cancel_reason, cancelled_at, created_at, updated_at
FROM rides
WHERE status = 'active' AND deleted_at IS NULL AND destination_address = ANY($1::text[])
ORDER BY departure_time ASC
LIMIT $2;
-- name: UpdateRideSeats :exec
UPDATE rides
SET available_seats = available_seats - $1,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: DeleteRide :exec
UPDATE rides
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateBooking :one
INSERT INTO bookings (
    id, ride_id, passenger_id, seats_booked, total_price, is_full_car_buyout, offered_price, status, pickup_preference, dropoff_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetBookingForUpdate :one
SELECT * FROM bookings WHERE id = $1 AND deleted_at IS NULL FOR UPDATE;

-- name: CountPassengerNoShows :one
SELECT COUNT(*) FROM bookings
WHERE passenger_id = $1 AND status = 'passenger_no_show' AND created_at > $2 AND deleted_at IS NULL;

-- name: GetBookingsByPassengerID :many
SELECT * FROM bookings WHERE passenger_id = $1 AND deleted_at IS NULL;

-- name: GetBookingByID :one
SELECT * FROM bookings WHERE id = $1 AND deleted_at IS NULL;


-- name: DeleteRidesByDriverID :exec
UPDATE rides SET deleted_at = CURRENT_TIMESTAMP
WHERE driver_id = $1 AND deleted_at IS NULL;

-- name: UpdateBookingStatus :exec
UPDATE bookings
SET status = $1,
    updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL;

-- name: UpdateBookingPriceAndStatus :exec
UPDATE bookings
SET status = $1,
    offered_price = $2,
    updated_at = NOW()
WHERE id = $3 AND deleted_at IS NULL;

-- name: DeleteBooking :exec
UPDATE bookings
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateRideStop :one
INSERT INTO ride_stops (id, ride_id, location, order_index)
VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5)
RETURNING *;

-- name: GetRideStopsByRideID :many
SELECT id, ride_id, ST_X(location::geometry) as lon, ST_Y(location::geometry) as lat, address, order_index
FROM ride_stops
WHERE ride_id = $1 AND deleted_at IS NULL
ORDER BY order_index ASC;

-- name: GetRidesByIDs :many
SELECT 
    id, driver_id, vehicle_id, vehicle_data, ST_X(origin::geometry) as origin_lon, ST_Y(origin::geometry) as origin_lat, ST_X(destination::geometry) as dest_lon, ST_Y(destination::geometry) as dest_lat, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, has_ac, has_luggage, takes_parcels, pickup_point, pickup_lat, pickup_lon, music_taste, no_smoking, no_pets, driving_speed, max_two_back, status, is_negotiable, has_fair_price_badge, allows_city_detour, city_detour_fee, cancel_reason, cancelled_at, created_at, updated_at
FROM rides
WHERE id = ANY(@ids::uuid[]) AND deleted_at IS NULL;

-- name: RejectExpiredNegotiations :many
UPDATE bookings
SET status = 'rejected', updated_at = NOW()
WHERE status = 'pending_negotiation' AND created_at < NOW() - INTERVAL '30 minutes' AND deleted_at IS NULL
RETURNING id, passenger_id, ride_id;



-- name: GetBookingsByRideID :many
SELECT * FROM bookings WHERE ride_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC;

-- name: GetRidesByDriverID :many
SELECT 
    id, driver_id, vehicle_id, vehicle_data, ST_X(origin::geometry) as origin_lon, ST_Y(origin::geometry) as origin_lat, ST_X(destination::geometry) as dest_lon, ST_Y(destination::geometry) as dest_lat, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, has_ac, has_luggage, takes_parcels, pickup_point, pickup_lat, pickup_lon, music_taste, no_smoking, no_pets, driving_speed, max_two_back, status, is_negotiable, has_fair_price_badge, allows_city_detour, city_detour_fee, cancel_reason, cancelled_at, created_at, updated_at
FROM rides
WHERE driver_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetRideStopsByRideIDs :many
SELECT id, ride_id, ST_X(location::geometry) as lon, ST_Y(location::geometry) as lat, address, order_index
FROM ride_stops
WHERE ride_id = ANY(@ride_ids::uuid[]) AND deleted_at IS NULL
ORDER BY order_index ASC;

-- name: GetDriverStatisticsAggregated :one
SELECT 
    COUNT(DISTINCT r.id)::int as total_rides,
    COUNT(DISTINCT CASE WHEN r.status = 'completed' THEN r.id END)::int as completed_rides,
    COUNT(DISTINCT CASE WHEN r.status = 'cancelled' THEN r.id END)::int as cancelled_rides,
    COALESCE(SUM(CASE WHEN r.status = 'completed' AND b.status = 'confirmed' THEN b.total_price ELSE 0 END), 0)::float8 as total_earnings
FROM rides r
LEFT JOIN bookings b ON r.id = b.ride_id AND b.deleted_at IS NULL
WHERE r.driver_id = $1 AND r.deleted_at IS NULL;

-- name: UpdateRideRoutePolygon :exec
UPDATE rides 
SET route_polygon = ST_Buffer(ST_GeomFromText($2, 4326)::geography, $3::double precision)::geometry, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTotalRidesCount :one
SELECT COUNT(*) FROM rides WHERE deleted_at IS NULL;

-- name: GetActiveRidesCount :one
SELECT COUNT(*) FROM rides WHERE status = 'active' AND deleted_at IS NULL;

-- name: GetRidesChartData :many
SELECT DATE(created_at)::text as date, COUNT(*)::bigint as count
FROM rides
WHERE created_at >= NOW() - INTERVAL '7 days' AND deleted_at IS NULL
GROUP BY DATE(created_at)
ORDER BY date ASC;

-- name: CountUserRidesAsDriver :one
SELECT COUNT(*) FROM rides
WHERE driver_id = $1 AND deleted_at IS NULL;

-- name: CountUserRidesAsPassenger :one
SELECT COUNT(*) FROM bookings
WHERE passenger_id = $1 AND deleted_at IS NULL;

-- name: GetPopularRoutes :many
SELECT origin_address, destination_address, count(*)::bigint as count 
FROM rides 
WHERE deleted_at IS NULL AND origin_address != '' AND destination_address != ''
GROUP BY origin_address, destination_address 
ORDER BY count DESC 
LIMIT 4;

-- name: GetPublicRidesByDriverID :many
SELECT 
    id, driver_id, origin_address, destination_address,
    ST_X(origin::geometry) as origin_lon, ST_Y(origin::geometry) as origin_lat,
    ST_X(destination::geometry) as dest_lon, ST_Y(destination::geometry) as dest_lat,
    departure_time, price_per_seat, status
FROM rides
WHERE driver_id = $1 AND deleted_at IS NULL
ORDER BY departure_time DESC
LIMIT $2;

