package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

type TodoReadRepository struct {
	txManager *db.TxManager
	schema    string
}

func NewTodoReadRepository(
	txManager *db.TxManager,
	schema string,
) *TodoReadRepository {
	return &TodoReadRepository{
		txManager: txManager,
		schema:    schema,
	}
}

func (r *TodoReadRepository) GetByID(ctx context.Context, id uuid.UUID) (*todo.Todo, error) {
	var result *todo.Todo
	err := r.txManager.WithinTransactionReadonly(ctx, r.schema, func(q *querydb.Queries) error {
		repo := NewTodoRepository(nil, q)
		var err error
		result, err = repo.GetByID(ctx, id)
		return err
	})
	return result, err
}

func (r *TodoReadRepository) List(ctx context.Context) ([]todo.Todo, error) {
	var result []todo.Todo
	err := r.txManager.WithinTransactionReadonly(ctx, r.schema, func(q *querydb.Queries) error {
		repo := NewTodoRepository(nil, q)
		var err error
		result, err = repo.List(ctx)
		return err
	})
	return result, err
}