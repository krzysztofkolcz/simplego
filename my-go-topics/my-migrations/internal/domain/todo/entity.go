package todo

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID        uuid.UUID
	Title     string
	Completed bool
	CreatedAt time.Time
}