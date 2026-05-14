package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
)

// OutboxWorker periodically reads unpublished events from each tenant's
// outbox_events table and marks them as published. In a real system this
// is where you'd forward events to a message broker (NATS, Kafka, etc.).
// Here we log them as a stand-in.
type OutboxWorker struct {
	txManager *db.TxManager
	publicQ   *commanddb.Queries
	interval  time.Duration
}

func NewOutboxWorker(txManager *db.TxManager, publicQ *commanddb.Queries, interval time.Duration) *OutboxWorker {
	return &OutboxWorker{
		txManager: txManager,
		publicQ:   publicQ,
		interval:  interval,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.ProcessOnce(ctx); err != nil {
				slog.ErrorContext(ctx, "outbox worker failed", "err", err)
			}
		}
	}
}

func (w *OutboxWorker) ProcessOnce(ctx context.Context) error {
	schemas, err := w.publicQ.ListTenantSchemas(ctx)
	if err != nil {
		return err
	}

	for _, schema := range schemas {
		if err := w.processTenant(ctx, schema); err != nil {
			slog.ErrorContext(ctx, "outbox: failed to process tenant",
				"schema", schema, "err", err)
		}
	}

	return nil
}

func (w *OutboxWorker) processTenant(ctx context.Context, schema string) error {
	return w.txManager.WithinTransaction(ctx, schema, func(q *commanddb.Queries) error {
		events, err := q.SelectUnpublishedOutboxEvents(ctx)
		if err != nil {
			return err
		}

		for _, e := range events {
			slog.InfoContext(ctx, "outbox: dispatching event",
				"schema", schema,
				"id", e.ID,
				"event_name", e.EventName,
				"payload", string(e.Payload),
			)
			if err := q.MarkOutboxEventPublished(ctx, e.ID); err != nil {
				return err
			}
		}

		return nil
	})
}
