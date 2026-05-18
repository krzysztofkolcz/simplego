package catalog

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, c Component) error
	GetByID(ctx context.Context, id uuid.UUID) (*Component, error)
	List(ctx context.Context) ([]Component, error)
}
