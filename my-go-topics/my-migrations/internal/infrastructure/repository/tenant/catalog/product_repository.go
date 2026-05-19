package catalog

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	domcatalog "github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
)

type ProductRepository struct {
	commandQ *commanddb.Queries
	queryQ   *querydb.Queries
}

func NewProductRepository(commandQ *commanddb.Queries, queryQ *querydb.Queries) *ProductRepository {
	return &ProductRepository{commandQ: commandQ, queryQ: queryQ}
}

func (r *ProductRepository) Create(ctx context.Context, p domcatalog.Product) error {
	_, err := r.commandQ.CreateProduct(ctx, commanddb.CreateProductParams{
		ID:          p.ID,
		Name:        p.Name,
		Description: nullableText(p.Description),
		Price:       float64ToNumeric(p.Price),
		Tags:        p.Tags,
	})
	if err != nil {
		return err
	}
	for _, item := range p.Recipe {
		if err := r.commandQ.UpsertProductComponent(ctx, commanddb.UpsertProductComponentParams{
			ProductID:   p.ID,
			ComponentID: item.ComponentID,
			Quantity:    int32(item.Quantity),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domcatalog.Product, error) {
	row, err := r.queryQ.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	components, err := r.queryQ.ListProductComponentsByProductID(ctx, id)
	if err != nil {
		return nil, err
	}
	return productFromQueryRow(row, components), nil
}

func (r *ProductRepository) Update(ctx context.Context, p domcatalog.Product) error {
	if err := r.commandQ.UpdateProduct(ctx, commanddb.UpdateProductParams{
		ID:          p.ID,
		Name:        p.Name,
		Description: nullableText(p.Description),
		Price:       float64ToNumeric(p.Price),
		Tags:        p.Tags,
	}); err != nil {
		return err
	}
	return r.syncRecipe(ctx, p)
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.commandQ.DeleteProduct(ctx, id)
}

// syncRecipe synchronizes product_components in DB to match p.Recipe.
// Deletes components removed from the recipe, upserts the rest.
func (r *ProductRepository) syncRecipe(ctx context.Context, p domcatalog.Product) error {
	existing, err := r.commandQ.ListProductComponentsByProductID(ctx, p.ID)
	if err != nil {
		return err
	}

	wanted := make(map[uuid.UUID]int, len(p.Recipe))
	for _, item := range p.Recipe {
		wanted[item.ComponentID] = item.Quantity
	}

	for _, row := range existing {
		if _, ok := wanted[row.ComponentID]; !ok {
			if err := r.commandQ.DeleteProductComponent(ctx, commanddb.DeleteProductComponentParams{
				ProductID:   p.ID,
				ComponentID: row.ComponentID,
			}); err != nil {
				return err
			}
		}
	}

	for _, item := range p.Recipe {
		if err := r.commandQ.UpsertProductComponent(ctx, commanddb.UpsertProductComponentParams{
			ProductID:   p.ID,
			ComponentID: item.ComponentID,
			Quantity:    int32(item.Quantity),
		}); err != nil {
			return err
		}
	}
	return nil
}
