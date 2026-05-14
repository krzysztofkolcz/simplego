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

	var t *todo.Todo
	err := h.uow.Execute(ctx, func(repo todo.Repository) error {
		var err error
		t, err = repo.GetByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		t.Delete()
		return repo.Delete(ctx, cmd.ID)
	})
	if err != nil {
		return err
	}

	return h.publisher.Publish(ctx, t.PullEvents())
}
