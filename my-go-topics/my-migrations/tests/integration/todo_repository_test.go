package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
	"github.com/stretchr/testify/require"
)

func TestTodoRepository_CreateAndGet(
	t *testing.T,
) {

	ctx := context.Background()

	txManager := db.NewTxManager(ConnectionPool)

	err = txManager.WithinTransaction(
		ctx,
		schema,
		func(commandQ *commanddb.Queries) error {

			queryQ := querydb.New(commandQ.DB)

			repo := tenantrepo.NewTodoRepository(
				commandQ,
				queryQ,
			)

			td := todo.Todo{
				ID:    uuid.New(),
				Title: "learn sqlc",
			}

			err := repo.Create(ctx, td)
			require.NoError(t, err)

			result, err := repo.GetByID(
				ctx,
				td.ID,
			)
			require.NoError(t, err)

			require.Equal(
				t,
				"learn sqlc",
				result.Title,
			)

			return nil
		},
	)

	require.NoError(t, err)
}
