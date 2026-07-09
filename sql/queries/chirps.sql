-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAllChirpsAsc :many
SELECT * FROM chirps
ORDER BY created_at ASC
LIMIT $1
OFFSET $2;

-- name: GetAllChirpsDesc :many
SELECT * FROM chirps
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1;

-- name: GetUserIdFromChirpId :one
SELECT user_id FROM chirps
WHERE id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1;

-- name: GetChirpsByUserIDAsc :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC
LIMIT $2
OFFSET $3;

-- name: GetChirpsByUserIDDesc :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;