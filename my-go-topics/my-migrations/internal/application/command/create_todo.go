package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type CreateTodoCommand struct {
	TenantSchema string
	Title        string
}

type CreateTodoHandler struct {
	uow       todo.UnitOfWork
	publisher domain.EventPublisher
}

func NewCreateTodoHandler(uow todo.UnitOfWork, publisher domain.EventPublisher) *CreateTodoHandler {
	return &CreateTodoHandler{uow: uow, publisher: publisher}
}

func (h *CreateTodoHandler) Handle(
	ctx context.Context,
	cmd CreateTodoCommand,
) (uuid.UUID, error) {

	t, err := todo.NewTodo(uuid.New(), cmd.Title)
	if err != nil {
		return uuid.Nil, err
	}

	err = h.uow.Execute(ctx, func(repo todo.Repository) error {
		return repo.Create(ctx, *t)
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Publish after successful transaction.
	// In production, replace with an OutboxPublisher that writes events
	// to an outbox table in the same transaction for guaranteed delivery.
	if err := h.publisher.Publish(ctx, t.PullEvents()); err != nil {
		return t.ID, err
	}

	return t.ID, nil
}
