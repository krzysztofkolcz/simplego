package domain

import "context"

type DomainEvent interface {
	EventName() string
}

type EventPublisher interface {
	Publish(ctx context.Context, events []DomainEvent) error
}
