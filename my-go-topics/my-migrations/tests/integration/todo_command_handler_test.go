package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
	tenantrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant"
)

func TestCreateTodoHandler_Handle(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)

	var createdID uuid.UUID

	err := txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)
		handler := command.NewCreateTodoHandler(repo)

		id, err := handler.Handle(ctx, command.CreateTodoCommand{Title: "buy milk"})
		if err != nil {
			return err
		}
		createdID = id
		return nil
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, createdID)

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

	var id uuid.UUID

	err := txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)
		handler := command.NewCreateTodoHandler(repo)

		var err error
		id, err = handler.Handle(ctx, command.CreateTodoCommand{Title: "read a book"})
		return err
	})
	require.NoError(t, err)

	err = txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)
		handler := command.NewCompleteTodoHandler(repo)

		return handler.Handle(ctx, command.CompleteTodoCommand{ID: id})
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

func TestDeleteTodoHandler_Handle(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)

	var id uuid.UUID

	err := txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)
		handler := command.NewCreateTodoHandler(repo)

		var err error
		id, err = handler.Handle(ctx, command.CreateTodoCommand{Title: "write tests"})
		return err
	})
	require.NoError(t, err)

	err = txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
		queryQ := querydb.New(commandQ.DB())
		repo := tenantrepo.NewTodoRepository(commandQ, queryQ)
		handler := command.NewDeleteTodoHandler(repo)

		return handler.Handle(ctx, command.DeleteTodoCommand{ID: id})
	})
	require.NoError(t, err)

	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
		repo := tenantrepo.NewTodoRepository(nil, queryQ)

		_, err := repo.GetByID(ctx, id)
		require.ErrorIs(t, err, pgx.ErrNoRows)

		return nil
	})
	require.NoError(t, err)
}
