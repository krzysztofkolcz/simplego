package catalog

import "context"

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(ctx context.Context, components ComponentRepository, products ProductRepository) error) error
}
