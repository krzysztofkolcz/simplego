package catalog

import (
	"context"

	"github.com/google/uuid"

	domcatalog "github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
)

type ComponentReadRepository struct {
	txManager *db.TxManager
	schema    string
}

func NewComponentReadRepository(txManager *db.TxManager, schema string) *ComponentReadRepository {
	return &ComponentReadRepository{txManager: txManager, schema: schema}
}

func (r *ComponentReadRepository) GetByID(ctx context.Context, id uuid.UUID) (*domcatalog.Component, error) {
	var result *domcatalog.Component
	err := r.txManager.WithinTransactionReadonly(ctx, r.schema, func(q *querydb.Queries) error {
		repo := NewComponentRepository(nil, q)
		var err error
		result, err = repo.GetByID(ctx, id)
		return err
	})
	return result, err
}

func (r *ComponentReadRepository) List(ctx context.Context) ([]domcatalog.Component, error) {
	var result []domcatalog.Component
	err := r.txManager.WithinTransactionReadonly(ctx, r.schema, func(q *querydb.Queries) error {
		rows, err := q.ListComponents(ctx)
		if err != nil {
			return err
		}
		result = make([]domcatalog.Component, len(rows))
		for i, row := range rows {
			result[i] = *componentFromQueryRow(row)
		}
		return nil
	})
	return result, err
}

type ProductReadRepository struct {
	txManager *db.TxManager
	schema    string
}

func NewProductReadRepository(txManager *db.TxManager, schema string) *ProductReadRepository {
	return &ProductReadRepository{txManager: txManager, schema: schema}
}

func (r *ProductReadRepository) GetByID(ctx context.Context, id uuid.UUID) (*domcatalog.Product, error) {
	var result *domcatalog.Product
	err := r.txManager.WithinTransactionReadonly(ctx, r.schema, func(q *querydb.Queries) error {
		repo := NewProductRepository(nil, q)
		var err error
		result, err = repo.GetByID(ctx, id)
		return err
	})
	return result, err
}

func (r *ProductReadRepository) List(ctx context.Context) ([]domcatalog.Product, error) {
	var result []domcatalog.Product
	err := r.txManager.WithinTransactionReadonly(ctx, r.schema, func(q *querydb.Queries) error {
		rows, err := q.ListProducts(ctx)
		if err != nil {
			return err
		}
		result = make([]domcatalog.Product, len(rows))
		for i, row := range rows {
			components, err := q.ListProductComponentsByProductID(ctx, row.ID)
			if err != nil {
				return err
			}
			result[i] = *productFromQueryRow(row, components)
		}
		return nil
	})
	return result, err
}
