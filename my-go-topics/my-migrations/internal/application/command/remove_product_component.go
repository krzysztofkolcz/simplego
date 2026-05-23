package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
)

type RemoveProductComponentCommand struct {
	TenantSchema string
	ProductID    uuid.UUID
	ComponentID  uuid.UUID
}

type RemoveProductComponentHandler struct {
	uow       catalog.UnitOfWork
	publisher domain.EventPublisher
}

func NewRemoveProductComponentHandler(
	uow       catalog.UnitOfWork,
	publisher domain.EventPublisher,
)*RemoveProductComponentHandler{
	return &RemoveProductComponentHandler{
		uow: uow,
		publisher: publisher,
	}
}

func (h *RemoveProductComponentHandler) Handle(ctx context.Context, cmd RemoveProductComponentCommand) error {
	return h.uow.Execute(ctx,
		func(txCtx context.Context, components catalog.ComponentRepository, products catalog.ProductRepository) error {
			product, err := products.GetByID(txCtx, cmd.ProductID)
			if err != nil {
				return err
			}
			err = product.RemoveComponent(cmd.ComponentID)
			if err != nil {
				return err
			}

			return products.Update(txCtx, *product)
	})
}