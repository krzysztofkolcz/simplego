package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

type TodoRepository struct {
	commandQ *commanddb.Queries
	queryQ   *querydb.Queries
}

func NewTodoRepository(
	commandQ *commanddb.Queries,
	queryQ *querydb.Queries,
) *TodoRepository {

	return &TodoRepository{
		commandQ: commandQ,
		queryQ:   queryQ,
	}
}

func (r *TodoRepository) Create(
	ctx context.Context,
	t todo.Todo,
) error {

	_, err := r.commandQ.CreateTodo(
		ctx,
		commanddb.CreateTodoParams{
			ID:    t.ID,
			Title: t.Title,
		},
	)

	return err
}

func (r *TodoRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*todo.Todo, error) {

	row, err := r.queryQ.GetTodo(ctx, id)
	if err != nil {
		return nil, err
	}

	return &todo.Todo{
		ID:        row.ID,
		Title:     row.Title,
		Completed: row.Completed,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *TodoRepository) Complete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.commandQ.CompleteTodo(ctx, id)
}

func (r *TodoRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.commandQ.DeleteTodo(ctx, id)
}