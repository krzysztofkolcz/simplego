package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testhelpers "github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestCreateTodoHandler_Handle_ReturnsValidID(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	id, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateTodoHandler_Handle_StoresTodo(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	id, _ := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.Len(t, repo.Created, 1)
	require.Equal(t, id, repo.Created[0].ID)
	require.Equal(t, "buy milk", repo.Created[0].Title)
}

func TestCreateTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{Err: errors.New("db down")}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	_, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.EqualError(t, err, "db down")
}

func TestCreateTodoHandler_Handle_UoWError(t *testing.T) {
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Err: errors.New("tx failed")})

	_, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.EqualError(t, err, "tx failed")
}
