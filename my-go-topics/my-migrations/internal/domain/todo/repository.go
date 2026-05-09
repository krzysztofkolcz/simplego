package todo

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, todo Todo) error
	GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	Complete(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}