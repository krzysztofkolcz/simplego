package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateUserHandler_Handle_ReturnsValidID(t *testing.T) {
	repo := &fakeUserRepo{}
	handler := NewCreateUserHandler(&fakeUserUoW{repo: repo})

	id, err := handler.Handle(context.Background(), CreateUserCommand{Email: "user@example.com"})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateUserHandler_Handle_StoresUser(t *testing.T) {
	repo := &fakeUserRepo{}
	handler := NewCreateUserHandler(&fakeUserUoW{repo: repo})

	id, _ := handler.Handle(context.Background(), CreateUserCommand{Email: "user@example.com"})

	require.Len(t, repo.created, 1)
	require.Equal(t, id, repo.created[0].ID)
	require.Equal(t, "user@example.com", repo.created[0].Email)
}

func TestCreateUserHandler_Handle_RepoError(t *testing.T) {
	repo := &fakeUserRepo{err: errors.New("db down")}
	handler := NewCreateUserHandler(&fakeUserUoW{repo: repo})

	_, err := handler.Handle(context.Background(), CreateUserCommand{Email: "user@example.com"})

	require.EqualError(t, err, "db down")
}
