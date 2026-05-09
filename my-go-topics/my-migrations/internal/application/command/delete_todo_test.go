package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteTodoHandler_Handle_CallsDelete(t *testing.T) {
	repo := &fakeTodoRepo{}
	handler := NewDeleteTodoHandler(&fakeTodoUoW{repo: repo})
	id := uuid.New()

	err := handler.Handle(context.Background(), DeleteTodoCommand{ID: id})

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{id}, repo.deleted)
}

func TestDeleteTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &fakeTodoRepo{err: errors.New("db down")}
	handler := NewDeleteTodoHandler(&fakeTodoUoW{repo: repo})

	err := handler.Handle(context.Background(), DeleteTodoCommand{ID: uuid.New()})

	require.EqualError(t, err, "db down")
}
