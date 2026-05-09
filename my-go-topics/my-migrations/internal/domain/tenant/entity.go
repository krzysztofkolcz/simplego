package tenant

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID
	SchemaName string
	CreatedAt time.Time
	MigrationStatus string
	MigrationError string
	MigrationUpdated time.Time
}