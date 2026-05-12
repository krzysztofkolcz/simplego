package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testhelpers "github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestCompleteTodoHandler_Handle_CallsComplete(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewCompleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})
	id := uuid.New()

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: id})

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{id}, repo.Completed)
}

func TestCompleteTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{Err: errors.New("db down")}
	handler := NewCompleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: uuid.New()})

	require.EqualError(t, err, "db down")
}
