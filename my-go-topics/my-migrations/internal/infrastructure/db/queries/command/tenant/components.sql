-- name: CreateComponent :one
INSERT INTO components (
    id,
    name,
    description,
    code,
    manufacturer,
    manufacturer_url,
    price,
    weight_g,
    quantity,
    image_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetComponentByID :one
SELECT * FROM components WHERE id = $1;

-- name: UpdateComponent :exec
UPDATE components SET
    name             = $2,
    description      = $3,
    manufacturer     = $4,
    manufacturer_url = $5,
    price            = $6,
    weight_g         = $7,
    quantity         = $8,
    image_url        = $9
WHERE id = $1;

-- name: DeleteComponent :exec
DELETE FROM components WHERE id = $1;
