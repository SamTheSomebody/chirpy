-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, NOW(), NOW(), $2, $3, NULL)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens WHERE revoked_at IS NULL AND token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE token = $1; 
