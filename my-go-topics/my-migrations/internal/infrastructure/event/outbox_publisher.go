package event

import (
	"context"
	"encoding/json"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
)

// OutboxPublisher writes domain events to the outbox_events table inside the
// active transaction carried by ctx. The outbox worker reads and dispatches
// them later, guaranteeing at-least-once delivery even if the process crashes.
type OutboxPublisher struct{}

func NewOutboxPublisher() *OutboxPublisher {
	return &OutboxPublisher{}
}

func (p *OutboxPublisher) Publish(ctx context.Context, events []domain.DomainEvent) error {
	q, ok := db.TxQueriesFromCtx(ctx)
	if !ok {
		return nil
	}
	for _, e := range events {
		payload, err := json.Marshal(e)
		if err != nil {
			return err
		}
		if err := q.InsertOutboxEvent(ctx, commanddb.InsertOutboxEventParams{
			EventName: e.EventName(),
			Payload:   payload,
		}); err != nil {
			return err
		}
	}
	return nil
}
