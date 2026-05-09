package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type DeleteTodoCommand struct {
	ID uuid.UUID
}

type DeleteTodoHandler struct {
	repo todo.Repository
}

func NewDeleteTodoHandler(repo todo.Repository) *DeleteTodoHandler {
	return &DeleteTodoHandler{repo: repo}
}

func (h *DeleteTodoHandler) Handle(
	ctx context.Context,
	cmd DeleteTodoCommand,
) error {

	return h.repo.Delete(ctx, cmd.ID)
}
