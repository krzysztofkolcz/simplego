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

	return h.uow.Execute(ctx, func(txCtx context.Context, repo todo.Repository) error {
		t, err := repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return err
		}
		if err := t.Complete(); err != nil {
			return err
		}
		if err := repo.Update(txCtx, *t); err != nil {
			return err
		}
		return h.publisher.Publish(txCtx, t.PullEvents())
	})
}
