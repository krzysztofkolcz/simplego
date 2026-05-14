package worker

import (
	"context"
	"log/slog"
	"time"

	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
)

// OutboxWorker periodically reads unpublished events from outbox_events and
// marks them as published. In a real system this is where you'd forward events
// to a message broker (NATS, Kafka, etc.). Here we log them as a stand-in.
type OutboxWorker struct {
	queries  *commanddb.Queries
	interval time.Duration
}

func NewOutboxWorker(queries *commanddb.Queries, interval time.Duration) *OutboxWorker {
	return &OutboxWorker{queries: queries, interval: interval}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processOnce(ctx); err != nil {
				slog.ErrorContext(ctx, "outbox worker failed", "err", err)
			}
		}
	}
}

func (w *OutboxWorker) processOnce(ctx context.Context) error {
	events, err := w.queries.SelectUnpublishedOutboxEvents(ctx)
	if err != nil {
		return err
	}

	for _, e := range events {
		slog.InfoContext(ctx, "outbox: dispatching event",
			"id", e.ID,
			"event_name", e.EventName,
			"payload", string(e.Payload),
		)
		if err := w.queries.MarkOutboxEventPublished(ctx, e.ID); err != nil {
			return err
		}
	}

	return nil
}
