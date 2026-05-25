package port

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
)

type CreateUserPort interface {
	Handle(ctx context.Context, cmd appcommand.CreateUserCommand) (uuid.UUID, error)
}

type GetUserPort interface {
	Handle(ctx context.Context, q appquery.GetUserQuery) (*appquery.GetUserResult, error)
}

type GetUserByEmailPort interface {
	Handle(ctx context.Context, q appquery.GetUserByEmailQuery) (*appquery.GetUserByEmailResult, error)
}
