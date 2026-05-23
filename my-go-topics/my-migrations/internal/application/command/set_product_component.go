package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
)

type SetProductComponentCommand struct {
	TenantSchema string
	ProductID    uuid.UUID
	ComponentID  uuid.UUID
	Quantity     int
}

type SetProductComponentHandler struct {
	uow       catalog.UnitOfWork
	publisher domain.EventPublisher
}

func NewSetProductComponentHandler(
	uow catalog.UnitOfWork,
	publisher domain.EventPublisher,
) *SetProductComponentHandler {
	return &SetProductComponentHandler{
		uow:       uow,
		publisher: publisher,
	}
}

func (h *SetProductComponentHandler) Handle(ctx context.Context, cmd SetProductComponentCommand) error {

	return h.uow.Execute(ctx,
		func(txCtx context.Context, components catalog.ComponentRepository, products catalog.ProductRepository) error {
			product, err := products.GetByID(txCtx, cmd.ProductID)
			if err != nil {
				return err
			}
			err = product.AddComponent(cmd.ComponentID, cmd.Quantity)
			if err != nil {
				return err
			}

			return products.Update(txCtx, *product)
		})
}
