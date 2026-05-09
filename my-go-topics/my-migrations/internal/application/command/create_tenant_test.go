package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateTenantHandler_Handle_ReturnsValidID(t *testing.T) {
	repo := &fakeTenantRepo{}
	handler := NewCreateTenantHandler(&fakeTenantUoW{repo: repo})

	id, err := handler.Handle(context.Background(), CreateTenantCommand{SchemaName: "acme"})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateTenantHandler_Handle_StoresTenant(t *testing.T) {
	repo := &fakeTenantRepo{}
	handler := NewCreateTenantHandler(&fakeTenantUoW{repo: repo})

	id, _ := handler.Handle(context.Background(), CreateTenantCommand{SchemaName: "acme"})

	require.Len(t, repo.created, 1)
	require.Equal(t, id, repo.created[0].ID)
	require.Equal(t, "acme", repo.created[0].SchemaName)
}

func TestCreateTenantHandler_Handle_RepoError(t *testing.T) {
	repo := &fakeTenantRepo{err: errors.New("db down")}
	handler := NewCreateTenantHandler(&fakeTenantUoW{repo: repo})

	_, err := handler.Handle(context.Background(), CreateTenantCommand{SchemaName: "acme"})

	require.EqualError(t, err, "db down")
}
