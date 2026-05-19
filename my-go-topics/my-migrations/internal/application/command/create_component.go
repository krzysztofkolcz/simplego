package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
)

type CreateComponentCommand struct {
	TenantSchema    string
	Name            string
	Code            string
	Description     string
	Manufacturer    string
	ManufacturerURL string
	Price           float64
	WeightG         int
	Quantity        int
	ImageURL        string
}

type CreateComponentHandler struct {
	uow       catalog.UnitOfWork
	publisher domain.EventPublisher
}

func NewCreateComponentHandler(uow catalog.UnitOfWork, publisher domain.EventPublisher) *CreateComponentHandler {
	return &CreateComponentHandler{uow: uow, publisher: publisher}
}

func (h *CreateComponentHandler) Handle(ctx context.Context, cmd CreateComponentCommand) (uuid.UUID, error) {
	c, err := catalog.NewComponent(uuid.New(), cmd.Name, cmd.Code)
	if err != nil {
		return uuid.Nil, err
	}
	c.Description = cmd.Description
	c.Manufacturer = cmd.Manufacturer
	c.ManufacturerURL = cmd.ManufacturerURL
	c.Price = cmd.Price
	c.WeightG = cmd.WeightG
	c.Quantity = cmd.Quantity
	c.ImageURL = cmd.ImageURL

	err = h.uow.Execute(ctx, 
		func(txCtx context.Context, components catalog.ComponentRepository, _ catalog.ProductRepository) error {
			if err := components.Create(txCtx, *c); err != nil {
				return err
			}
			return h.publisher.Publish(txCtx, c.PullEvents())
		},
	)
	if err != nil {
		return uuid.Nil, err
	}
	return c.ID, nil
}
