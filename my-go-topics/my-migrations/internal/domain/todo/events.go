package todo

import (
	"time"

	"github.com/google/uuid"
)

type TodoCreated struct {
	TodoID     uuid.UUID
	Title      string
	OccurredAt time.Time
}

func (e TodoCreated) EventName() string { return "todo.created" }

type TodoCompleted struct {
	TodoID     uuid.UUID
	OccurredAt time.Time
}

func (e TodoCompleted) EventName() string { return "todo.completed" }

type TodoDeleted struct {
	TodoID     uuid.UUID
	OccurredAt time.Time
}

func (e TodoDeleted) EventName() string { return "todo.deleted" }
