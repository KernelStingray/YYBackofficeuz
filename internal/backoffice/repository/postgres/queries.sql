-- name: CreateAdmin :one
INSERT INTO admins (id, login, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAdminByLogin :one
SELECT * FROM admins
WHERE login = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ListAdmins :many
SELECT * FROM admins
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpsertRemoteConfig :one
INSERT INTO remote_configs (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE 
SET value = EXCLUDED.value, updated_at = NOW(), deleted_at = NULL
RETURNING *;

-- name: GetRemoteConfig :one
SELECT * FROM remote_configs
WHERE key = $1 AND deleted_at IS NULL LIMIT 1;

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (id, admin_id, action, entity, target_id, payload, ip_address)
VALUES ($1, $2, $3, $4, $5, $6, $7);


