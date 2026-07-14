package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/catalog"
	"github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func ptr[T any](v T) *T { return &v }

func newCatalogUoW(components *testhelpers.FakeCatalogComponentRepo) *testhelpers.FakeCatalogUoW {
	return &testhelpers.FakeCatalogUoW{
		Components: components,
		Products:   &testhelpers.FakeCatalogProductRepo{},
	}
}

func seedComponent(repo *testhelpers.FakeCatalogComponentRepo) catalog.Component {
	c, _ := catalog.NewComponent(uuid.New(), "Resistor 10kΩ", "RES-10K")
	c.Description = "original description"
	c.Manufacturer = "Yageo"
	c.ManufacturerURL = "https://yageo.com"
	c.Price = 0.05
	c.WeightG = 1
	c.Quantity = 100
	c.ImageURL = "https://img.example.com/res.png"
	repo.Store = append(repo.Store, *c)
	return *c
}

func TestUpdateComponentHandler_UpdatesOnlyProvidedFields(t *testing.T) {
	repo := &testhelpers.FakeCatalogComponentRepo{}
	original := seedComponent(repo)
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(newCatalogUoW(repo), pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{
		ID:          original.ID,
		Description: ptr("new description"),
		Price:       ptr(0.10),
	})

	require.NoError(t, err)
	updated := repo.Store[0]
	require.Equal(t, "new description", updated.Description)
	require.Equal(t, 0.10, updated.Price)
	// untouched fields stay as-is
	require.Equal(t, original.Manufacturer, updated.Manufacturer)
	require.Equal(t, original.ManufacturerURL, updated.ManufacturerURL)
	require.Equal(t, original.WeightG, updated.WeightG)
	require.Equal(t, original.Quantity, updated.Quantity)
	require.Equal(t, original.ImageURL, updated.ImageURL)
}

func TestUpdateComponentHandler_UpdatesAllFields(t *testing.T) {
	repo := &testhelpers.FakeCatalogComponentRepo{}
	original := seedComponent(repo)
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(newCatalogUoW(repo), pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{
		ID:              original.ID,
		Description:     ptr("updated desc"),
		Manufacturer:    ptr("Murata"),
		ManufacturerURL: ptr("https://murata.com"),
		Price:           ptr(0.20),
		WeightG:         ptr(2),
		Quantity:        ptr(50),
		ImageURL:        ptr("https://img.example.com/new.png"),
	})

	require.NoError(t, err)
	updated := repo.Store[0]
	require.Equal(t, "updated desc", updated.Description)
	require.Equal(t, "Murata", updated.Manufacturer)
	require.Equal(t, "https://murata.com", updated.ManufacturerURL)
	require.Equal(t, 0.20, updated.Price)
	require.Equal(t, 2, updated.WeightG)
	require.Equal(t, 50, updated.Quantity)
	require.Equal(t, "https://img.example.com/new.png", updated.ImageURL)
}

func TestUpdateComponentHandler_SavesUpdatedComponent(t *testing.T) {
	repo := &testhelpers.FakeCatalogComponentRepo{}
	original := seedComponent(repo)
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(newCatalogUoW(repo), pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{
		ID:   original.ID,
		Price: ptr(9.99),
	})

	require.NoError(t, err)
	require.Len(t, repo.Updated, 1)
	require.Equal(t, original.ID, repo.Updated[0].ID)
}

func TestUpdateComponentHandler_PublishesComponentUpdatedEvent(t *testing.T) {
	repo := &testhelpers.FakeCatalogComponentRepo{}
	original := seedComponent(repo)
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(newCatalogUoW(repo), pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{
		ID:    original.ID,
		Price: ptr(1.23),
	})

	require.NoError(t, err)
	require.Len(t, pub.Published, 1)

	event, ok := pub.Published[0].(catalog.ComponentUpdated)
	require.True(t, ok)
	require.Equal(t, original.ID, event.ComponentID)
	require.NotZero(t, event.OccurredAt)
}

func TestUpdateComponentHandler_NilFields_DoNotChangeComponent(t *testing.T) {
	repo := &testhelpers.FakeCatalogComponentRepo{}
	original := seedComponent(repo)
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(newCatalogUoW(repo), pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{ID: original.ID})

	require.NoError(t, err)
	updated := repo.Store[0]
	require.Equal(t, original.Description, updated.Description)
	require.Equal(t, original.Manufacturer, updated.Manufacturer)
	require.Equal(t, original.Price, updated.Price)
	require.Equal(t, original.Quantity, updated.Quantity)
}

func TestUpdateComponentHandler_NotFound(t *testing.T) {
	repo := &testhelpers.FakeCatalogComponentRepo{}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(newCatalogUoW(repo), pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{ID: uuid.New()})

	require.ErrorIs(t, err, domain.ErrNotFound)
	require.Empty(t, pub.Published)
}

func TestUpdateComponentHandler_RepoGetByIDError(t *testing.T) {
	repo := &testhelpers.FakeCatalogComponentRepo{Err: errors.New("db down")}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(newCatalogUoW(repo), pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{ID: uuid.New()})

	require.EqualError(t, err, "db down")
	require.Empty(t, pub.Published)
}

func TestUpdateComponentHandler_UoWError(t *testing.T) {
	uow := &testhelpers.FakeCatalogUoW{Err: errors.New("tx failed")}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewUpdateComponentHandler(uow, pub)

	err := handler.Handle(context.Background(), UpdateComponentCommand{ID: uuid.New()})

	require.EqualError(t, err, "tx failed")
	require.Empty(t, pub.Published)
}
