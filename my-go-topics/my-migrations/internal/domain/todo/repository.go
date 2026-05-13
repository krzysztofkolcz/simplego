package todo

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, todo Todo) error
	GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	Update(ctx context.Context, todo Todo) error
	Delete(ctx context.Context, id uuid.UUID) error
}
