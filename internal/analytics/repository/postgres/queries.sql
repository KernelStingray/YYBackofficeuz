-- name: CreateEvent :one
INSERT INTO events (id, event_type, user_id, payload)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEventStats :many
SELECT event_type, COUNT(*) as event_count
FROM events
WHERE deleted_at IS NULL
GROUP BY event_type
ORDER BY event_count DESC;
