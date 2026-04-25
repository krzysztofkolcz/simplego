-- name: CreateTodo :exec
-- Inserts a new todo row. The ID is supplied by the application (pre-generated
-- in the HTTP handler), not by the database default, so we know it before
-- executing the query.
INSERT INTO todos (id, title, description, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetTodoByID :one
SELECT id, title, description, status, created_at, updated_at
FROM todos
WHERE id = $1;

-- name: ListTodos :many
-- Returns all todos ordered by creation time (newest first).
SELECT id, title, description, status, created_at, updated_at
FROM todos
ORDER BY created_at DESC;

-- name: UpdateTodo :exec
-- Updates only title, description, and updated_at.
-- Status is intentionally excluded — it is managed by CompleteTodo below.
-- Having two separate queries mirrors the CQRS command split on the
-- application side: UpdateTodoCommand ≠ CompleteTodoCommand.
UPDATE todos
SET
    title       = $2,
    description = $3,
    updated_at  = $4
WHERE id = $1;

-- name: CompleteTodo :exec
-- Sets status = 'done' for a single todo.
-- This query only touches status + updated_at, making it impossible to
-- accidentally change the title during a Complete operation.
UPDATE todos
SET
    status     = 'done',
    updated_at = $2
WHERE id = $1;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = $1;
