package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
	tenantrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant"
)

func TestTodoRepository_Create(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)

	var createdID uuid.UUID

	err := txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)

		td := todo.Todo{
			ID:    uuid.New(),
			Title: "learn sqlc",
		}
		createdID = td.ID

		return repo.Create(ctx, td)
	})
	require.NoError(t, err)

	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
		repo := tenantrepo.NewTodoRepository(nil, queryQ)

		result, err := repo.GetByID(ctx, createdID)
		require.NoError(t, err)
		require.Equal(t, "learn sqlc", result.Title)
		require.False(t, result.Completed)

		return nil
	})
	require.NoError(t, err)
}

func TestTodoRepository_Complete(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)

	id := uuid.New()

	err := txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)

		return repo.Create(ctx, todo.Todo{ID: id, Title: "to be completed"})
	})
	require.NoError(t, err)

	err = txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)

		t, err := repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if err := t.Complete(); err != nil {
			return err
		}
		return repo.Update(ctx, *t)
	})
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

func TestTodoRepository_Delete(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)

	id := uuid.New()

	err := txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)

		return repo.Create(ctx, todo.Todo{ID: id, Title: "to be deleted"})
	})
	require.NoError(t, err)

	err = txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)

		return repo.Delete(ctx, id)
	})
	require.NoError(t, err)

	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
		repo := tenantrepo.NewTodoRepository(nil, queryQ)

		_, err := repo.GetByID(ctx, id)
		require.ErrorIs(t, err, domain.ErrNotFound)

		return nil
	})
	require.NoError(t, err)
}
