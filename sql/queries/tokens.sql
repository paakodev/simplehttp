-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, token, user_id, expires_at, revoked_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshTokenByUserID :one
SELECT * FROM refresh_tokens 
WHERE user_id = $1;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens 
WHERE token = $1;

-- name: GetUserIDByRefreshToken :one
SELECT * FROM refresh_tokens 
WHERE token = $1;

-- name: ExtendTokenRevoke :exec
UPDATE refresh_tokens 
SET expires_at = $1 
WHERE token = $2;

-- name: RevokeRefreshTokenByUserID :exec
UPDATE refresh_tokens 
SET revoked_at = NOW() 
WHERE user_id = $1;

-- name: RevokeRefreshTokenByToken :exec
UPDATE refresh_tokens 
SET revoked_at = NOW() 
WHERE token = $1;

-- name: DeleteRefreshTokenByUserID :exec
DELETE FROM refresh_tokens 
WHERE user_id = $1;

-- name: DeleteRefreshTokenByToken :exec
DELETE FROM refresh_tokens 
WHERE token = $1;