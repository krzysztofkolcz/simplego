package port

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
)

type CreateComponentPort interface {
	Handle(ctx context.Context, cmd appcommand.CreateComponentCommand) (uuid.UUID, error)
}

type GetComponentPort interface {
	Handle(ctx context.Context, q appquery.GetComponentQuery) (*appquery.GetComponentResult, error)
}

type CreateProductPort interface {
	Handle(ctx context.Context, cmd appcommand.CreateProductCommand) (uuid.UUID, error)
}

type GetProductPort interface {
	Handle(ctx context.Context, q appquery.GetProductQuery) (*appquery.GetProductResult, error)
}
