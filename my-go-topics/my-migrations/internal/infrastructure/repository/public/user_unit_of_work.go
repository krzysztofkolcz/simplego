package public

import (
	"context"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

type UserUnitOfWork struct {
	txManager *db.TxManager
}

func NewUserUnitOfWork(txManager *db.TxManager) *UserUnitOfWork {
	return &UserUnitOfWork{txManager: txManager}
}

func (u *UserUnitOfWork) Execute(
	ctx context.Context,
	fn func(repo user.Repository) error,
) error {

	return u.txManager.WithinPublicTransaction(
		ctx,
		func(commandQ *commanddb.Queries) error {
			queryQ := querydb.New(commandQ.DB())
			repo := NewUserRepository(commandQ, queryQ)
			return fn(repo)
		},
	)
}
