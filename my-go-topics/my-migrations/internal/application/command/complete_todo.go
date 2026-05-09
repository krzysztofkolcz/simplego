package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type CompleteTodoCommand struct {
	ID uuid.UUID
}

type CompleteTodoHandler struct {
	repo todo.Repository
}

func NewCompleteTodoHandler(repo todo.Repository) *CompleteTodoHandler {
	return &CompleteTodoHandler{repo: repo}
}

func (h *CompleteTodoHandler) Handle(
	ctx context.Context,
	cmd CompleteTodoCommand,
) error {

	return h.repo.Complete(ctx, cmd.ID)
}
