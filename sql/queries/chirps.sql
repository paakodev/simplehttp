-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC
LIMIT $1
OFFSET $2;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1;