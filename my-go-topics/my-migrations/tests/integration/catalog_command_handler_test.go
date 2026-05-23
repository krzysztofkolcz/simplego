package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
	infraevent "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/event"
	catalogrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant/catalog"
)

func createComponentHelper(t *testing.T, ctx context.Context, txManager *db.TxManager) uuid.UUID {
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewCreateComponentHandler(uow, infraevent.NewLogPublisher())

	id, err := handler.Handle(ctx, command.CreateComponentCommand{
		Name:         "Przedłużka z blaszką wykończeniową typu łezka*srebro AG 925*R1 50 38 mm - Metal : AG 925 - bez pokrycia",
		Code:         uuid.New().String(),
		Description:  "",
		Manufacturer: "Silvexcraft",
		Price:        12.16,
		WeightG:      4,
		Quantity:     1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)
	return id
}

func createProductHelper(t *testing.T, ctx context.Context, txManager *db.TxManager) uuid.UUID{
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewCreateProductHandler(uow, infraevent.NewLogPublisher())

	id, err := handler.Handle(ctx, command.CreateProductCommand{
		Name:        "Bransoletka z ametystu",
		Description: "Bransoletka z ametystu",
		Price:       129.99,
		Tags:        []string{"bransoletka", "ametyst", "srebro"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)
	return id
}

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

// TestSetProductComponentHandler_AddNew — dodanie nowego komponentu do produktu, odczyt z DB weryfikuje, że receptura zawiera komponent z oczekiwaną ilością
func TestSetProductComponentHandler_AddNew(t *testing.T){
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	productID := createProductHelper(t,ctx,txManager)
	componentID := createComponentHelper(t, ctx, txManager)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	readRepo := catalogrepo.NewProductReadRepository(txManager, TenantSchema)
	handler := command.NewSetProductComponentHandler(uow, infraevent.NewLogPublisher())
	command := command.SetProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    productID,
		ComponentID:  componentID,
		Quantity:     1,
	}
	err := handler.Handle(ctx, command)
	require.NoError(t, err)

	product, err := readRepo.GetByID(ctx, productID)
	require.NoError(t, err)
	require.Equal(t, 1, len(product.Recipe))
	recipe := product.Recipe[0]
	require.Equal(t, componentID, recipe.ComponentID)
	require.Equal(t, 1, recipe.Quantity)
}
func TestSetProductComponentHandler_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	productID := createProductHelper(t, ctx, txManager)
	componentID := createComponentHelper(t, ctx, txManager)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	readRepo := catalogrepo.NewProductReadRepository(txManager, TenantSchema)
	handler := command.NewSetProductComponentHandler(uow, infraevent.NewLogPublisher())
	cmd := command.SetProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    productID,
		ComponentID:  componentID,
		Quantity:     1,
	}
	require.NoError(t, handler.Handle(ctx, cmd))
	cmd.Quantity = 5
	require.NoError(t, handler.Handle(ctx, cmd))

	product, err := readRepo.GetByID(ctx, productID)
	require.NoError(t, err)
	require.Len(t, product.Recipe, 1)
	require.Equal(t, 5, product.Recipe[0].Quantity)
}

func TestSetProductComponentHandler_InvalidQuantity(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	productID := createProductHelper(t, ctx, txManager)
	componentID := createComponentHelper(t, ctx, txManager)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewSetProductComponentHandler(uow, infraevent.NewLogPublisher())

	err := handler.Handle(ctx, command.SetProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    productID,
		ComponentID:  componentID,
		Quantity:     0,
	})
	require.ErrorIs(t, err, catalog.ErrInvalidQuantity)
}

func TestSetProductComponentHandler_ProductNotFound(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewSetProductComponentHandler(uow, infraevent.NewLogPublisher())

	err := handler.Handle(ctx, command.SetProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    uuid.New(),
		ComponentID:  uuid.New(),
		Quantity:     1,
	})
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRemoveProductComponentHandler_OK(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	productID := createProductHelper(t, ctx, txManager)
	componentID := createComponentHelper(t, ctx, txManager)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	readRepo := catalogrepo.NewProductReadRepository(txManager, TenantSchema)

	setHandler := command.NewSetProductComponentHandler(uow, infraevent.NewLogPublisher())
	require.NoError(t, setHandler.Handle(ctx, command.SetProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    productID,
		ComponentID:  componentID,
		Quantity:     2,
	}))

	removeHandler := command.NewRemoveProductComponentHandler(uow, infraevent.NewLogPublisher())
	require.NoError(t, removeHandler.Handle(ctx, command.RemoveProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    productID,
		ComponentID:  componentID,
	}))

	product, err := readRepo.GetByID(ctx, productID)
	require.NoError(t, err)
	require.Empty(t, product.Recipe)
}

func TestRemoveProductComponentHandler_ComponentNotInRecipe(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	productID := createProductHelper(t, ctx, txManager)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewRemoveProductComponentHandler(uow, infraevent.NewLogPublisher())

	err := handler.Handle(ctx, command.RemoveProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    productID,
		ComponentID:  uuid.New(),
	})
	require.ErrorIs(t, err, catalog.ErrComponentNotFound)
}

func TestRemoveProductComponentHandler_ProductNotFound(t *testing.T) {
	ctx := context.Background()
	txManager := db.NewTxManager(ConnectionPool)
	uow := catalogrepo.NewCatalogUnitOfWork(txManager, TenantSchema)
	handler := command.NewRemoveProductComponentHandler(uow, infraevent.NewLogPublisher())

	err := handler.Handle(ctx, command.RemoveProductComponentCommand{
		TenantSchema: TenantSchema,
		ProductID:    uuid.New(),
		ComponentID:  uuid.New(),
	})
	require.ErrorIs(t, err, domain.ErrNotFound)
}