-- name: CreateProduct :one
INSERT INTO products (
    id,
    name,
    description,
    price,
    tags
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: UpdateProduct :exec
UPDATE products SET
    name        = $2,
    description = $3,
    price       = $4,
    tags        = $5
WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: UpsertProductComponent :exec
INSERT INTO product_components (product_id, component_id, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (product_id, component_id) DO UPDATE SET quantity = EXCLUDED.quantity;

-- name: DeleteProductComponent :exec
DELETE FROM product_components
WHERE product_id = $1 AND component_id = $2;

-- name: ListProductComponentsByProductID :many
SELECT product_id, component_id, quantity
FROM product_components
WHERE product_id = $1;
