package catalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"

	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
)

type ComponentRepository struct {
	commandQ *commanddb.Queries
	queryQ   *querydb.Queries
}

func NewComponentRepository (
	commandQ *commanddb.Queries,
	queryQ *querydb.Queries,
) *ComponentRepository {

	return &ComponentRepository{
		commandQ: commandQ,
		queryQ:   queryQ,
	}
}

func (r *ComponentRepository) Create(ctx context.Context, c catalog.Component) error {
	panic("not implemented") // TODO: Implement

}

func (r *ComponentRepository) GetByID(ctx context.Context, id uuid.UUID) (*catalog.Component, error) {
	panic("not implemented") // TODO: Implement
}

func (r *ComponentRepository) List(ctx context.Context) ([]catalog.Component, error) {
	panic("not implemented") // TODO: Implement
}