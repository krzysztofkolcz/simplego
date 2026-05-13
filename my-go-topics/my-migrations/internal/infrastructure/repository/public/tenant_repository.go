package public

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/tenant"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

type TenantRepository struct {
	commandQ *commanddb.Queries
	queryQ   *querydb.Queries
}

func NewTenantRepository(
	commandQ *commanddb.Queries,
	queryQ *querydb.Queries,
) *TenantRepository {

	return &TenantRepository{
		commandQ: commandQ,
		queryQ:   queryQ,
	}
}

func (r *TenantRepository) Create(
	ctx context.Context,
	t tenant.Tenant,
) error {

	_, err := r.commandQ.CreateTenant(
		ctx,
		commanddb.CreateTenantParams{
			ID:         t.ID,
			SchemaName: t.SchemaName,
		},
	)

	return err
}

func (r *TenantRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*tenant.Tenant, error) {

	row, err := r.queryQ.GetTenantByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &tenant.Tenant{
		ID:               row.ID,
		SchemaName:       row.SchemaName,
		CreatedAt:        row.CreatedAt.Time,
		MigrationStatus:  row.MigrationStatus.String,
		MigrationError:   row.MigrationError.String,
		MigrationUpdated: row.MigrationUpdatedAt.Time,
	}, nil
}
