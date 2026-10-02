-- name: GetTrustScore :one
SELECT * FROM trust_scores
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: InitTrustScore :one
INSERT INTO trust_scores (user_id, score)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateTrustScore :one
UPDATE trust_scores
SET score = $2, updated_at = NOW()
WHERE user_id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CreateFraudEvent :one
INSERT INTO fraud_events (id, user_id, event_type, description, penalty_applied)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFraudEventsByUser :many
SELECT * FROM fraud_events
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;
