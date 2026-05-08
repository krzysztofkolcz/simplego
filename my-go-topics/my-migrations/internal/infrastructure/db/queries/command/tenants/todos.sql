-- name: CreateTodo :one
INSERT INTO todos (
    id,
    title
)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: CompleteTodo :exec
UPDATE todos
SET completed = true
WHERE id = $1;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1;