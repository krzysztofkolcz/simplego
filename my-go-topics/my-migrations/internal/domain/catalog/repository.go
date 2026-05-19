package catalog

import (
	"context"

	"github.com/google/uuid"
)

type ComponentRepository interface {
	Create(ctx context.Context, c Component) error
	GetByID(ctx context.Context, id uuid.UUID) (*Component, error)
	Update(ctx context.Context, c Component) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProductRepository interface {
	Create(ctx context.Context, p Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Update(ctx context.Context, p Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}
