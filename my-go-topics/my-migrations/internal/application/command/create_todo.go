package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type CreateTodoCommand struct {
	Title string
}

type CreateTodoHandler struct {
	repo todo.Repository
}

func NewCreateTodoHandler(repo todo.Repository) *CreateTodoHandler {
	return &CreateTodoHandler{repo: repo}
}

func (h *CreateTodoHandler) Handle(
	ctx context.Context,
	cmd CreateTodoCommand,
) (uuid.UUID, error) {

	t := todo.Todo{
		ID:    uuid.New(),
		Title: cmd.Title,
	}

	if err := h.repo.Create(ctx, t); err != nil {
		return uuid.Nil, err
	}

	return t.ID, nil
}
