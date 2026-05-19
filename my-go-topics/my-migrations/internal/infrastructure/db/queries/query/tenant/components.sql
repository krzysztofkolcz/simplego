-- name: GetComponentByID :one
SELECT * FROM components WHERE id = $1;

-- name: ListComponents :many
SELECT * FROM components ORDER BY name;
