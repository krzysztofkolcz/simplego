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

type ComponentRepository struct {
	commandQ *commanddb.Queries
	queryQ   *querydb.Queries
}

func NewComponentRepository(
	commandQ *commanddb.Queries,
	queryQ *querydb.Queries,
) *ComponentRepository {
	return &ComponentRepository{
		commandQ: commandQ,
		queryQ:   queryQ,
	}
}

func (r *ComponentRepository) Create(ctx context.Context, c domcatalog.Component) error {
	_, err := r.commandQ.CreateComponent(ctx, commanddb.CreateComponentParams{
		ID:              c.ID,
		Name:            c.Name,
		Code:            c.Code,
		Description:     nullableText(c.Description),
		Manufacturer:    nullableText(c.Manufacturer),
		ManufacturerUrl: nullableText(c.ManufacturerURL),
		Price:           float64ToNumeric(c.Price),
		WeightG:         nullableInt4(c.WeightG),
		Quantity:        int32(c.Quantity),
		ImageUrl:        nullableText(c.ImageURL),
	})
	return err
}

func (r *ComponentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domcatalog.Component, error) {
	row, err := r.queryQ.GetComponentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return componentFromQueryRow(row), nil
}

func (r *ComponentRepository) Update(ctx context.Context, c domcatalog.Component) error {
	return r.commandQ.UpdateComponent(ctx, commanddb.UpdateComponentParams{
		ID:              c.ID,
		Name:            c.Name,
		Description:     nullableText(c.Description),
		Manufacturer:    nullableText(c.Manufacturer),
		ManufacturerUrl: nullableText(c.ManufacturerURL),
		Price:           float64ToNumeric(c.Price),
		WeightG:         nullableInt4(c.WeightG),
		Quantity:        int32(c.Quantity),
		ImageUrl:        nullableText(c.ImageURL),
	})
}

func (r *ComponentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.commandQ.DeleteComponent(ctx, id)
}
