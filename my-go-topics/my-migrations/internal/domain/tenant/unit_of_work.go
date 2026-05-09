package tenant

import "context"

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(repo Repository) error) error
}
