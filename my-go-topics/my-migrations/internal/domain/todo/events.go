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
