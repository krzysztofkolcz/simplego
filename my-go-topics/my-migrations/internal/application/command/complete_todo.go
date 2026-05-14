package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type CompleteTodoCommand struct {
	TenantSchema string
	ID           uuid.UUID
}

type CompleteTodoHandler struct {
	uow       todo.UnitOfWork
	publisher domain.EventPublisher
}

func NewCompleteTodoHandler(uow todo.UnitOfWork, publisher domain.EventPublisher) *CompleteTodoHandler {
	return &CompleteTodoHandler{uow: uow, publisher: publisher}
}

func (h *CompleteTodoHandler) Handle(
	ctx context.Context,
	cmd CompleteTodoCommand,
) error {

	var t *todo.Todo
	err := h.uow.Execute(ctx, func(repo todo.Repository) error {
		var err error
		t, err = repo.GetByID(ctx, cmd.ID)
		if err != nil {
			return err
		}
		if err := t.Complete(); err != nil {
			return err
		}
		return repo.Update(ctx, *t)
	})
	if err != nil {
		return err
	}

	return h.publisher.Publish(ctx, t.PullEvents())
}
