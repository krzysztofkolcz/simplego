package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, t *Todo) error
	GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	List(ctx context.Context) ([]*Todo, error)
}