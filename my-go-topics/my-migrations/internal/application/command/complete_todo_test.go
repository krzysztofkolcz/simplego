package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCompleteTodoHandler_Handle_CallsComplete(t *testing.T) {
	repo := &fakeTodoRepo{}
	handler := NewCompleteTodoHandler(&fakeTodoUoW{repo: repo})
	id := uuid.New()

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: id})

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{id}, repo.completed)
}

func TestCompleteTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &fakeTodoRepo{err: errors.New("db down")}
	handler := NewCompleteTodoHandler(&fakeTodoUoW{repo: repo})

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: uuid.New()})

	require.EqualError(t, err, "db down")
}
