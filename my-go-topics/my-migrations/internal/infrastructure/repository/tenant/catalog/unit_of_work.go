package catalog

import (
	"context"

	domcatalog "github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
)

type CatalogUnitOfWork struct {
	txManager    *db.TxManager
	tenantSchema string
}

func NewCatalogUnitOfWork(txManager *db.TxManager, tenantSchema string) *CatalogUnitOfWork {
	return &CatalogUnitOfWork{txManager: txManager, tenantSchema: tenantSchema}
}

func (u *CatalogUnitOfWork) Execute(
	ctx context.Context,
	fn func(ctx context.Context, components domcatalog.ComponentRepository, products domcatalog.ProductRepository) error,
) error {
	return u.txManager.WithinTransaction(ctx, u.tenantSchema, func(commandQ *commanddb.Queries) error {
		txCtx := db.ContextWithTxQueries(ctx, commandQ)
		queryQ := querydb.New(commandQ.DB())
		components := NewComponentRepository(commandQ, queryQ)
		products := NewProductRepository(commandQ, queryQ)
		return fn(txCtx, components, products)
	})
}
