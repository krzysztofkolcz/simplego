package catalog

import (
	"context"

	"github.com/google/uuid"
)

type ReadRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Catalog, error)
	List(ctx context.Context) ([]Catalog, error)
}
