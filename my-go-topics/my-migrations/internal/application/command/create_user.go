package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
)

type CreateUserCommand struct {
	Email string
}

type CreateUserHandler struct {
	uow user.UnitOfWork
}

func NewCreateUserHandler(uow user.UnitOfWork) *CreateUserHandler {
	return &CreateUserHandler{uow: uow}
}

func (h *CreateUserHandler) Handle(
	ctx context.Context,
	cmd CreateUserCommand,
) (uuid.UUID, error) {

	u := user.User{
		ID:    uuid.New(),
		Email: cmd.Email,
	}

	err := h.uow.Execute(ctx, func(repo user.Repository) error {
		return repo.Create(ctx, u)
	})
	if err != nil {
		return uuid.Nil, err
	}

	return u.ID, nil
}
