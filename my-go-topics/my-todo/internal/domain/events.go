package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the marker interface for all domain events.
// Events are named in the past tense — they describe something that happened.
// They are published after a command completes successfully.
//
// In a production system these would be delivered to a message broker
// (Kafka, NATS, etc.). Here we use an in-memory publisher that simply logs
// each event — enough to make the pattern visible without infrastructure noise.
type DomainEvent interface {
	// EventName returns a stable, dot-separated identifier for the event type.
	// Convention: "<aggregate>.<past-tense-verb>"
	EventName() string
}

// EventPublisher is the second PORT in the domain layer.
// Command handlers call Publish after every successful state change.
// The infrastructure layer provides the concrete implementation.
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
}

// --- Event types ---

// TodoCreated is published when a new Todo is successfully persisted.
type TodoCreated struct {
	TodoID     uuid.UUID
	Title      string
	OccurredAt time.Time
}

func (e TodoCreated) EventName() string { return "todo.created" }

// TodoCompleted is published when a Todo transitions to the "done" state.
type TodoCompleted struct {
	TodoID     uuid.UUID
	OccurredAt time.Time
}

func (e TodoCompleted) EventName() string { return "todo.completed" }

// TodoUpdated is published when a Todo's title or description is changed.
type TodoUpdated struct {
	TodoID     uuid.UUID
	NewTitle   string
	OccurredAt time.Time
}

func (e TodoUpdated) EventName() string { return "todo.updated" }

// TodoDeleted is published when a Todo is permanently removed.
type TodoDeleted struct {
	TodoID     uuid.UUID
	OccurredAt time.Time
}

func (e TodoDeleted) EventName() string { return "todo.deleted" }
