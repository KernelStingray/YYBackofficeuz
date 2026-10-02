-- name: CreateTicket :one
INSERT INTO tickets (id, user_id, subject, description, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserTickets :many
SELECT * FROM tickets
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateTicketStatus :one
UPDATE tickets
SET status = $2, resolution = $3, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTicket :exec
UPDATE tickets SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;
