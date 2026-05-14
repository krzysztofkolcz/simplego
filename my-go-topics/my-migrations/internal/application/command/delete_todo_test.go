package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	testhelpers "github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestDeleteTodoHandler_Handle_DeletesTodo(t *testing.T) {
	id := uuid.New()
	repo := &testhelpers.FakeTodoRepo{
		Created: []todo.Todo{{ID: id, Title: "buy milk"}},
	}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewDeleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

	err := handler.Handle(context.Background(), DeleteTodoCommand{ID: id})

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{id}, repo.Deleted)
	require.Len(t, pub.Published, 1)
	require.Equal(t, "todo.deleted", pub.Published[0].EventName())
}

func TestDeleteTodoHandler_Handle_NotFound(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewDeleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, &testhelpers.FakeEventPublisher{})

	err := handler.Handle(context.Background(), DeleteTodoCommand{ID: uuid.New()})

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeleteTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{Err: errors.New("db down")}
	handler := NewDeleteTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, &testhelpers.FakeEventPublisher{})

	err := handler.Handle(context.Background(), DeleteTodoCommand{ID: uuid.New()})

	require.EqualError(t, err, "db down")
}
