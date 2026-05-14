package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type DeleteTodoCommand struct {
	TenantSchema string
	ID           uuid.UUID
}

type DeleteTodoHandler struct {
	uow       todo.UnitOfWork
	publisher domain.EventPublisher
}

func NewDeleteTodoHandler(uow todo.UnitOfWork, publisher domain.EventPublisher) *DeleteTodoHandler {
	return &DeleteTodoHandler{uow: uow, publisher: publisher}
}

func (h *DeleteTodoHandler) Handle(
	ctx context.Context,
	cmd DeleteTodoCommand,
) error {

	return h.uow.Execute(ctx, func(txCtx context.Context, repo todo.Repository) error {
		t, err := repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return err
		}
		t.Delete()
		if err := repo.Delete(txCtx, cmd.ID); err != nil {
			return err
		}
		return h.publisher.Publish(txCtx, t.PullEvents())
	})
}
