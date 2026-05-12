package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testhelpers "github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestCreateUserHandler_Handle_ReturnsValidID(t *testing.T) {
	repo := &testhelpers.FakeUserRepo{}
	handler := NewCreateUserHandler(&testhelpers.FakeUserUoW{Repo: repo})

	id, err := handler.Handle(context.Background(), CreateUserCommand{Email: "user@example.com"})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateUserHandler_Handle_StoresUser(t *testing.T) {
	repo := &testhelpers.FakeUserRepo{}
	handler := NewCreateUserHandler(&testhelpers.FakeUserUoW{Repo: repo})

	id, _ := handler.Handle(context.Background(), CreateUserCommand{Email: "user@example.com"})

	require.Len(t, repo.Created, 1)
	require.Equal(t, id, repo.Created[0].ID)
	require.Equal(t, "user@example.com", repo.Created[0].Email)
}

func TestCreateUserHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeUserRepo{Err: errors.New("db down")}
	handler := NewCreateUserHandler(&testhelpers.FakeUserUoW{Repo: repo})

	_, err := handler.Handle(context.Background(), CreateUserCommand{Email: "user@example.com"})

	require.EqualError(t, err, "db down")
}
