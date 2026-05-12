package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testhelpers "github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestCreateTenantHandler_Handle_ReturnsValidID(t *testing.T) {
	repo := &testhelpers.FakeTenantRepo{}
	handler := NewCreateTenantHandler(&testhelpers.FakeTenantUoW{Repo: repo})

	id, err := handler.Handle(context.Background(), CreateTenantCommand{SchemaName: "acme"})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateTenantHandler_Handle_StoresTenant(t *testing.T) {
	repo := &testhelpers.FakeTenantRepo{}
	handler := NewCreateTenantHandler(&testhelpers.FakeTenantUoW{Repo: repo})

	id, _ := handler.Handle(context.Background(), CreateTenantCommand{SchemaName: "acme"})

	require.Len(t, repo.Created, 1)
	require.Equal(t, id, repo.Created[0].ID)
	require.Equal(t, "acme", repo.Created[0].SchemaName)
}

func TestCreateTenantHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeTenantRepo{Err: errors.New("db down")}
	handler := NewCreateTenantHandler(&testhelpers.FakeTenantUoW{Repo: repo})

	_, err := handler.Handle(context.Background(), CreateTenantCommand{SchemaName: "acme"})

	require.EqualError(t, err, "db down")
}
