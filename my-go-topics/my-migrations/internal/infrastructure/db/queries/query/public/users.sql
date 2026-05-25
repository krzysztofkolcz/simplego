-- name: GetUserID :one
SELECT
    id,
    email
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT
    id,
    email
FROM users
WHERE email = $1;