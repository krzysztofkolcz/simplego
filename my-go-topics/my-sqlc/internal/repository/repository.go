package repository

import (
	"context"

	"github.com/krzysztofkolcz/my-sqlc/internal/db"
)

type OrderRepo struct {
	q *db.Queries
}

func NewOrderRepo(q *db.Queries) *OrderRepo {
	return &OrderRepo{q: q}
}

func (r *OrderRepo) Create(ctx context.Context, params db.CreateOrderParams) error {
	return r.q.CreateOrder(ctx, params)
}
