package tenant

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, tenant Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
}