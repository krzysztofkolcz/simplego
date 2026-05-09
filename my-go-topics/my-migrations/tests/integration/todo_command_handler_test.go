package integration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
	tenantrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant"
)

func TestCreateTodoHandler_Handle(t *testing.T) {
	ctx := context.Background()
	uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
	handler := command.NewCreateTodoHandler(uow)

	createdID, err := handler.Handle(ctx, command.CreateTodoCommand{Title: "buy milk"})
	require.NoError(t, err)
	require.NotEmpty(t, createdID)

	txManager := db.NewTxManager(ConnectionPool)
	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
		repo := tenantrepo.NewTodoRepository(nil, queryQ)

		result, err := repo.GetByID(ctx, createdID)
		require.NoError(t, err)
		require.Equal(t, "buy milk", result.Title)
		require.False(t, result.Completed)

		return nil
	})
	require.NoError(t, err)
}

func TestCompleteTodoHandler_Handle(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	uow := tenantrepo.NewTodoUnitOfWork(txManager, TenantSchema)

	id, err := command.NewCreateTodoHandler(uow).Handle(ctx, command.CreateTodoCommand{Title: "read a book"})
	require.NoError(t, err)

	err = command.NewCompleteTodoHandler(uow).Handle(ctx, command.CompleteTodoCommand{ID: id})
	require.NoError(t, err)

	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
		repo := tenantrepo.NewTodoRepository(nil, queryQ)

		result, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.True(t, result.Completed)

		return nil
	})
	require.NoError(t, err)
}

func TestDeleteTodoHandler_Handle(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	uow := tenantrepo.NewTodoUnitOfWork(txManager, TenantSchema)

	id, err := command.NewCreateTodoHandler(uow).Handle(ctx, command.CreateTodoCommand{Title: "write tests"})
	require.NoError(t, err)

	err = command.NewDeleteTodoHandler(uow).Handle(ctx, command.DeleteTodoCommand{ID: id})
	require.NoError(t, err)

	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
		repo := tenantrepo.NewTodoRepository(nil, queryQ)

		_, err := repo.GetByID(ctx, id)
		require.ErrorIs(t, err, pgx.ErrNoRows)

		return nil
	})
	require.NoError(t, err)
}
