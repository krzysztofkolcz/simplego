package catalog

import (
	"context"

	"github.com/google/uuid"
)

type ComponentReadRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Component, error)
	List(ctx context.Context) ([]Component, error)
}

type ProductReadRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	List(ctx context.Context) ([]Product, error)
}
