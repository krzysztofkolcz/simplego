package repository

import (
	"context"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
)

type TodoUnitOfWork struct {
	txManager    *db.TxManager
	tenantSchema string
}

func NewTodoUnitOfWork(
	txManager *db.TxManager,
	tenantSchema string,
) *TodoUnitOfWork {

	return &TodoUnitOfWork{
		txManager:    txManager,
		tenantSchema: tenantSchema,
	}
}

func (u *TodoUnitOfWork) Execute(
	ctx context.Context,
	fn func(ctx context.Context, repo todo.Repository) error,
) error {

	return u.txManager.WithinTransaction(
		ctx,
		u.tenantSchema,
		func(commandQ *commanddb.Queries) error {
			txCtx := db.ContextWithTxQueries(ctx, commandQ)
			queryQ := querydb.New(commandQ.DB())
			repo := NewTodoRepository(commandQ, queryQ)
			return fn(txCtx, repo)
		},
	)
}
