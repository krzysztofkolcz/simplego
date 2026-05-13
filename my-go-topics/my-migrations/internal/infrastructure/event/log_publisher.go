package event

import (
	"context"
	"log/slog"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
)

type LogPublisher struct{}

func NewLogPublisher() *LogPublisher {
	return &LogPublisher{}
}

func (p *LogPublisher) Publish(ctx context.Context, events []domain.DomainEvent) error {
	for _, e := range events {
		slog.InfoContext(ctx, "domain event published", "event", e.EventName())
	}
	return nil
}
