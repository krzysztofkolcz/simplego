-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: ListProducts :many
SELECT * FROM products ORDER BY name;

-- name: ListProductComponentsByProductID :many
SELECT product_id, component_id, quantity
FROM product_components
WHERE product_id = $1;
