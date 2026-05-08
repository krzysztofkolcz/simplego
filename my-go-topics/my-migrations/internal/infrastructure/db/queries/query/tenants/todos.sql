-- name: GetTodo :one
SELECT
    id,
    title,
    completed,
    created_at
FROM todos
WHERE id = $1;

-- name: ListTodos :many
SELECT
    id,
    title,
    completed,
    created_at
FROM todos
ORDER BY created_at DESC;

-- name: ListIncompleteTodos :many
SELECT
    id,
    title,
    completed,
    created_at
FROM todos
WHERE completed = false
ORDER BY created_at DESC;