package port

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
)

type CreateTenantPort interface {
	Handle(ctx context.Context, cmd appcommand.CreateTenantCommand) (uuid.UUID, error)
}

type GetTenantPort interface {
	Handle(ctx context.Context, q appquery.GetTenantQuery) (*appquery.GetTenantResult, error)
}
