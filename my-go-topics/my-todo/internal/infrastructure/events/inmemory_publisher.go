// Package events provides the in-memory implementation of domain.EventPublisher.
//
// In a production system this would be replaced by a Kafka/NATS publisher.
// The domain.EventPublisher interface makes it swappable — main.go just wires
// a different implementation without touching any business logic.
//
// For this learning project, publishing simply logs the event with slog.
// This makes the event flow visible in the server output without requiring
// message broker infrastructure.
package events

import (
	"context"
	"log/slog"

	"github.com/C5383717/my-todo/internal/domain"
	slogctx "github.com/veqryn/slog-context"
)

// InMemoryEventPublisher implements domain.EventPublisher by logging events.
// It satisfies the port — swap it with a real publisher without changing any
// domain or application code.
type InMemoryEventPublisher struct{}

func NewInMemoryEventPublisher() *InMemoryEventPublisher {
	return &InMemoryEventPublisher{}
}

// Publish logs the domain event. In production, this would send the event
// to a message broker so other services can react to it.
func (p *InMemoryEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	slogctx.Info(ctx, "domain event published",
		slog.String("event", event.EventName()),
	)

	return nil
}
