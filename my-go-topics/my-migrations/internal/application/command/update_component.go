package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
)

type UpdateComponentCommand struct {
	TenantSchema    string
	ID              uuid.UUID
	Description     *string
	Manufacturer    *string
	ManufacturerURL *string
	Price           *float64
	WeightG         *int
	Quantity        *int
	ImageURL        *string
}

type UpdateComponentHandler struct {
	uow       catalog.UnitOfWork
	publisher domain.EventPublisher
}

func NewUpdateComponentHandler(uow catalog.UnitOfWork, publisher domain.EventPublisher) *UpdateComponentHandler {
	return &UpdateComponentHandler{uow: uow, publisher: publisher}
}

func (h *UpdateComponentHandler) Handle(ctx context.Context, cmd UpdateComponentCommand) error {
	return h.uow.Execute(ctx,
		func(txCtx context.Context, components catalog.ComponentRepository, _ catalog.ProductRepository) error {
			c, err := components.GetByID(txCtx, cmd.ID)
			if err != nil {
				return err
			}

			c.UpdateDetails(
				cmd.Description,
				cmd.Manufacturer,
				cmd.ManufacturerURL,
				cmd.Price,
				cmd.WeightG,
				cmd.Quantity,
				cmd.ImageURL,
			)

			if err := components.Update(txCtx, *c); err != nil {
				return err
			}
			return h.publisher.Publish(txCtx, c.PullEvents())
		},
	)
}
