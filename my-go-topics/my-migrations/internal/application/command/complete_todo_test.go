package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestCompleteTodoHandler_Handle_CompletesTodo(t *testing.T) {
	id := uuid.New()
	repo := &testhelpers.FakeTodoRepo{
		Created: []todo.Todo{{ID: id, Title: "buy milk"}},
	}
	handler := NewCompleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: id})

	require.NoError(t, err)
	require.True(t, repo.Created[0].Completed)
}

func TestCompleteTodoHandler_Handle_AlreadyCompleted(t *testing.T) {
	id := uuid.New()
	repo := &testhelpers.FakeTodoRepo{
		Created: []todo.Todo{{ID: id, Title: "buy milk", Completed: true}},
	}
	handler := NewCompleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: id})

	require.ErrorIs(t, err, domain.ErrAlreadyCompleted)
}

func TestCompleteTodoHandler_Handle_NotFound(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewCompleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: uuid.New()})

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCompleteTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{Err: errors.New("db down")}
	handler := NewCompleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo})

	err := handler.Handle(context.Background(), CompleteTodoCommand{ID: uuid.New()})

	require.EqualError(t, err, "db down")
}
