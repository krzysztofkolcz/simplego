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
	uow todo.UnitOfWork
}

func NewCompleteTodoHandler(uow todo.UnitOfWork) *CompleteTodoHandler {
	return &CompleteTodoHandler{uow: uow}
}

func (h *CompleteTodoHandler) Handle(
	ctx context.Context,
	cmd CompleteTodoCommand,
) error {

	return h.uow.Execute(ctx, func(repo todo.Repository) error {
		return repo.Complete(ctx, cmd.ID)
	})
}
