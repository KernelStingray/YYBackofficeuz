-- name: CreateMessage :one
INSERT INTO chats_messages (
    id, sender_id, receiver_id, ride_id, reply_to_id, content, status
) VALUES (
    $1, $2, $3, $4, $5, $6, 'sent'
) RETURNING *;

-- name: GetChatHistory :many
SELECT * FROM chats_messages
WHERE 
    ((sender_id = $1 AND receiver_id = $2) 
    OR (sender_id = $2 AND receiver_id = $1))
    AND ($3::uuid IS NULL OR ride_id = $3)
    AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT 100;

-- name: MarkMessageAsRead :exec
UPDATE chats_messages
SET status = 'read', updated_at = NOW()
WHERE id = $1 AND receiver_id = $2 AND deleted_at IS NULL;

-- name: DeleteMessage :exec
UPDATE chats_messages SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;

-- name: UpsertDeviceToken :exec
INSERT INTO device_tokens (user_id, fcm_token, platform, language, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (user_id) DO UPDATE 
SET fcm_token = EXCLUDED.fcm_token,
    platform = EXCLUDED.platform,
    language = EXCLUDED.language,
    updated_at = NOW(),
    deleted_at = NULL;

-- name: GetDeviceToken :one
SELECT fcm_token, platform, language FROM device_tokens
WHERE user_id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: DeleteDeviceToken :exec
UPDATE device_tokens SET deleted_at = NOW() WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CreateNotification :one
INSERT INTO notifications (
    id, user_id, title, body, type, is_read, ride_id, booking_id, parcel_id, data
) VALUES (
    $1, $2, $3, $4, $5, FALSE, $6, $7, $8, $9
) RETURNING *;

-- name: GetNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET is_read = TRUE, updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
