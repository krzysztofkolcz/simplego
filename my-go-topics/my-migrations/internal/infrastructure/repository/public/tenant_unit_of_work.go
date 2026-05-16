package public

import (
	"context"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/tenant"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/public"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/public"
)

type TenantUnitOfWork struct {
	txManager *db.TxManager
}

func NewTenantUnitOfWork(txManager *db.TxManager) *TenantUnitOfWork {
	return &TenantUnitOfWork{txManager: txManager}
}

func (u *TenantUnitOfWork) Execute(
	ctx context.Context,
	fn func(repo tenant.Repository) error,
) error {

	return u.txManager.WithinPublicTransaction(
		ctx,
		func(commandQ *commanddb.Queries) error {
			queryQ := querydb.New(commandQ.DB())
			repo := NewTenantRepository(commandQ, queryQ)
			return fn(repo)
		},
	)
}
