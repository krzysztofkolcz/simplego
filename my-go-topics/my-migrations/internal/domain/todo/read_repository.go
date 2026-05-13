package todo

import (
	"context"

	"github.com/google/uuid"
)

type ReadRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	List(ctx context.Context) ([]Todo, error)
}