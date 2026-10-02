-- name: CreateUser :one
INSERT INTO users (first_name, last_name, gender, phone_number, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone_number = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetAllUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC;

-- name: UpdateUserStatus :one
UPDATE users SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: UpdateUserKYC :one
UPDATE users SET kyc_status = $2, first_name = $3, last_name = $4, kyc_rejection_reason = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: UpdateUserKYCStatusAndRole :one
UPDATE users SET role = $2, kyc_status = $3, kyc_rejection_reason = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users SET first_name = $2, last_name = $3, gender = $4, locale = COALESCE($5, locale), avatar_url = COALESCE($6, avatar_url), updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: AddTrustedDevice :exec
UPDATE users SET trusted_devices = array_append(trusted_devices, $2)
WHERE id = $1 AND NOT ($2 = ANY(trusted_devices));

-- name: CreateCar :one
INSERT INTO cars (
    user_id, gov_number, brand, model,
    tech_passport_photo, driver_license_photo,
    tech_passport_front, tech_passport_back,
    driver_license_front, driver_license_back,
    car_front_photo, car_back_photo, car_side_photo,
    color, year
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetCarsByUserID :many
SELECT * FROM cars WHERE user_id = $1 AND deleted_at IS NULL;

-- name: GetCarByID :one
SELECT * FROM cars WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: DeleteCar :exec
UPDATE cars SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateCar :one
UPDATE cars
SET gov_number = COALESCE(NULLIF(sqlc.arg('gov_number')::text, ''), gov_number),
    brand = COALESCE(NULLIF(sqlc.arg('brand')::text, ''), brand),
    model = COALESCE(NULLIF(sqlc.arg('model')::text, ''), model),
    color = COALESCE(NULLIF(sqlc.arg('color')::text, ''), color),
    year = COALESCE(NULLIF(sqlc.arg('year')::int, 0), year),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCarsModerationStatusByUserID :exec
UPDATE cars SET moderation_status = $2, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND deleted_at IS NULL;

-- name: UpdateUserRating :exec
UPDATE users 
SET rating = $2,
    rating_count = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: BlockUser :exec
INSERT INTO user_blocks (blocker_id, blocked_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnblockUser :exec
DELETE FROM user_blocks
WHERE blocker_id = $1 AND blocked_id = $2;

-- name: GetBlockedUsers :many
SELECT blocked_id
FROM user_blocks
WHERE blocker_id = $1;

-- name: CreateReport :exec
INSERT INTO reports (id, reporter_id, reported_id, reason)
VALUES ($1, $2, $3, $4);

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW(), phone = CONCAT('deleted_', id, '_', phone)
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTotalUsersCount :one
SELECT count(*) FROM users WHERE deleted_at IS NULL;

-- name: GetTotalDriversCount :one
SELECT count(*) FROM users WHERE role = 'driver' AND deleted_at IS NULL;

-- name: GetTotalPassengersCount :one
SELECT count(*) FROM users WHERE role = 'passenger' AND deleted_at IS NULL;

-- name: GetUsersChartData :many
SELECT DATE(created_at)::text as date, count(*)::bigint as count 
FROM users 
WHERE created_at >= NOW() - INTERVAL '7 days' AND deleted_at IS NULL
GROUP BY DATE(created_at) 
ORDER BY date ASC;

-- name: SearchUsersAdmin :many
SELECT * FROM users 
WHERE deleted_at IS NULL
  AND (@search::text = '' OR first_name ILIKE '%' || @search::text || '%' OR last_name ILIKE '%' || @search::text || '%' OR phone_number ILIKE '%' || @search::text || '%')
  AND (@role::text = '' OR role = @role::text)
  AND (@status::text = '' OR status = @status::text)
  AND (@kyc_status::text = '' OR kyc_status = @kyc_status::text)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountSearchUsersAdmin :one
SELECT count(*) FROM users 
WHERE deleted_at IS NULL
  AND (@search::text = '' OR first_name ILIKE '%' || @search::text || '%' OR last_name ILIKE '%' || @search::text || '%' OR phone_number ILIKE '%' || @search::text || '%')
  AND (@role::text = '' OR role = @role::text)
  AND (@status::text = '' OR status = @status::text)
  AND (@kyc_status::text = '' OR kyc_status = @kyc_status::text);
