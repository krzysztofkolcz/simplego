package catalog

import (
	"time"

	"github.com/google/uuid"
)

type ComponentUpdated struct {
	ComponentID uuid.UUID
	OccurredAt  time.Time
}

func (e ComponentUpdated) EventName() string { return "catalog.component.updated" }
