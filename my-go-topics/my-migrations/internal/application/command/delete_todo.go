package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type DeleteTodoCommand struct {
	TenantSchema string
	ID           uuid.UUID
}

type DeleteTodoHandler struct {
	uow todo.UnitOfWork
}

func NewDeleteTodoHandler(uow todo.UnitOfWork) *DeleteTodoHandler {
	return &DeleteTodoHandler{uow: uow}
}

func (h *DeleteTodoHandler) Handle(
	ctx context.Context,
	cmd DeleteTodoCommand,
) error {

	return h.uow.Execute(ctx, func(repo todo.Repository) error {
		return repo.Delete(ctx, cmd.ID)
	})
}
