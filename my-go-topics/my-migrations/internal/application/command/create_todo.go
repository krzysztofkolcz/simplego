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
	uow todo.UnitOfWork
}

func NewCreateTodoHandler(uow todo.UnitOfWork) *CreateTodoHandler {
	return &CreateTodoHandler{uow: uow}
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

	return t.ID, nil
}
