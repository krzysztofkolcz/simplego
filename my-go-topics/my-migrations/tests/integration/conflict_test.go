package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	publicrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/public"
)

func TestCreateTenantHandler_ConflictOnDuplicateSchema(t *testing.T) {
	ctx := context.Background()
	uow := publicrepo.NewTenantUnitOfWork(db.NewTxManager(ConnectionPool))

	_, err := command.NewCreateTenantHandler(uow).Handle(ctx, command.CreateTenantCommand{
		SchemaName: "duplicate_schema",
	})
	require.NoError(t, err)

	_, err = command.NewCreateTenantHandler(uow).Handle(ctx, command.CreateTenantCommand{
		SchemaName: "duplicate_schema",
	})
	require.ErrorIs(t, err, domain.ErrConflict)
}

func TestCreateUserHandler_ConflictOnDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	uow := publicrepo.NewUserUnitOfWork(db.NewTxManager(ConnectionPool))

	_, err := command.NewCreateUserHandler(uow).Handle(ctx, command.CreateUserCommand{
		Email: "duplicate@example.com",
	})
	require.NoError(t, err)

	_, err = command.NewCreateUserHandler(uow).Handle(ctx, command.CreateUserCommand{
		Email: "duplicate@example.com",
	})
	require.ErrorIs(t, err, domain.ErrConflict)
}
