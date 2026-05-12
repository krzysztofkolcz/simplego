package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testhelpers "github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestDeleteTodoHandler_Handle_CallsDelete(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewDeleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})
	id := uuid.New()

	err := handler.Handle(context.Background(), DeleteTodoCommand{ID: id})

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{id}, repo.Deleted)
}

func TestDeleteTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{Err: errors.New("db down")}
	handler := NewDeleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	err := handler.Handle(context.Background(), DeleteTodoCommand{ID: uuid.New()})

	require.EqualError(t, err, "db down")
}
