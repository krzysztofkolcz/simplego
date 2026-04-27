package persistence

import (
	"context"

	"github.com/krzysztofkolcz/my-sqlc/internal/db"
)

type UserRepository struct {
    q *db.Queries
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    u, err := r.q.GetUserByEmail(ctx, email)
    if err != nil {
        return nil, err
    }

    return &domain.User{
        ID: u.ID,
        Email: u.Email,
    }, nil
}