CREATE TABLE outbox_events (
    id          BIGSERIAL PRIMARY KEY,
    event_name  TEXT NOT NULL,
    payload     JSONB NOT NULL,
    created_at  TIMESTAMP DEFAULT now(),
    published_at TIMESTAMP
);
