package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/tenant"
)

type CreateTenantCommand struct {
	SchemaName string
}

type CreateTenantHandler struct {
	uow tenant.UnitOfWork
}

func NewCreateTenantHandler(uow tenant.UnitOfWork) *CreateTenantHandler {
	return &CreateTenantHandler{uow: uow}
}

func (h *CreateTenantHandler) Handle(
	ctx context.Context,
	cmd CreateTenantCommand,
) (uuid.UUID, error) {

	t := tenant.Tenant{
		ID:         uuid.New(),
		SchemaName: cmd.SchemaName,
	}

	err := h.uow.Execute(ctx, func(repo tenant.Repository) error {
		return repo.Create(ctx, t)
	})
	if err != nil {
		return uuid.Nil, err
	}

	return t.ID, nil
}
