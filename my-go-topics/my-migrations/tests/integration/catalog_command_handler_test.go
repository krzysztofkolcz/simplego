package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
	infraevent "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/event"
	catalogrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant/catalog"
)

func TestCreateComponentHandler_Handle(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewCreateComponentHandler(uow, infraevent.NewLogPublisher())

	id, err := handler.Handle(ctx, command.CreateComponentCommand{
		Name:         "Resistor 10kΩ",
		Code:         "RES-10K",
		Description:  "Standard resistor",
		Manufacturer: "Yageo",
		Price:        0.05,
		WeightG:      1,
		Quantity:     100,
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(q *querydb.Queries) error {
		repo := catalogrepo.NewComponentRepository(nil, q)

		c, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, "Resistor 10kΩ", c.Name)
		require.Equal(t, "RES-10K", c.Code)
		require.Equal(t, "Standard resistor", c.Description)
		require.Equal(t, "Yageo", c.Manufacturer)
		require.InDelta(t, 0.05, c.Price, 0.001)
		require.Equal(t, 1, c.WeightG)
		require.Equal(t, 100, c.Quantity)

		return nil
	})
	require.NoError(t, err)
}

func TestCreateProductHandler_Handle(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewCreateProductHandler(uow, infraevent.NewLogPublisher())

	id, err := handler.Handle(ctx, command.CreateProductCommand{
		Name:        "Widget Pro",
		Description: "A professional widget",
		Price:       29.99,
		Tags:        []string{"pro", "widget"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	err = txManager.WithinTransactionReadonly(ctx, TenantSchema, func(q *querydb.Queries) error {
		repo := catalogrepo.NewProductRepository(nil, q)

		p, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, "Widget Pro", p.Name)
		require.Equal(t, "A professional widget", p.Description)
		require.InDelta(t, 29.99, p.Price, 0.001)
		require.ElementsMatch(t, []string{"pro", "widget"}, p.Tags)
		require.Empty(t, p.Recipe)

		return nil
	})
	require.NoError(t, err)
}
