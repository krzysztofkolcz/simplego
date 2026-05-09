package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type GetTodoQuery struct {
	ID uuid.UUID
}

type GetTodoResult struct {
	ID        uuid.UUID
	Title     string
	Completed bool
}

type GetTodoHandler struct {
	repo todo.Repository
}

func NewGetTodoHandler(repo todo.Repository) *GetTodoHandler {
	return &GetTodoHandler{repo: repo}
}

func (h *GetTodoHandler) Handle(
	ctx context.Context,
	q GetTodoQuery,
) (*GetTodoResult, error) {

	t, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &GetTodoResult{
		ID:        t.ID,
		Title:     t.Title,
		Completed: t.Completed,
	}, nil
}
