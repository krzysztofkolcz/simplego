package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
)

type CreateProductCommand struct {
	TenantSchema string
	Name         string
	Description  string
	Price        float64
	Tags         []string
}

type CreateProductHandler struct {
	uow       catalog.UnitOfWork
	publisher domain.EventPublisher
}

func NewCreateProductHandler(uow catalog.UnitOfWork, publisher domain.EventPublisher) *CreateProductHandler {
	return &CreateProductHandler{uow: uow, publisher: publisher}
}

func (h *CreateProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (uuid.UUID, error) {
	p, err := catalog.NewProduct(uuid.New(), cmd.Name)
	if err != nil {
		return uuid.Nil, err
	}
	p.Description = cmd.Description
	p.Price = cmd.Price
	if len(cmd.Tags) > 0 {
		p.Tags = cmd.Tags
	}

	err = h.uow.Execute(ctx, func(txCtx context.Context, _ catalog.ComponentRepository, products catalog.ProductRepository) error {
		if err := products.Create(txCtx, *p); err != nil {
			return err
		}
		return h.publisher.Publish(txCtx, p.PullEvents())
	})
	if err != nil {
		return uuid.Nil, err
	}
	return p.ID, nil
}
