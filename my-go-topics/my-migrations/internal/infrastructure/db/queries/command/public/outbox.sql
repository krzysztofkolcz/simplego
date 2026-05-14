-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (event_name, payload)
VALUES ($1, $2);

-- name: SelectUnpublishedOutboxEvents :many
SELECT id, event_name, payload, created_at
FROM outbox_events
WHERE published_at IS NULL
ORDER BY id
LIMIT 100;

-- name: MarkOutboxEventPublished :exec
UPDATE outbox_events
SET published_at = now()
WHERE id = $1;
