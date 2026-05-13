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

func TestCreateTodoHandler_Handle_ReturnsValidID(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

	id, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateTodoHandler_Handle_StoresTodo(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

	id, _ := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.Len(t, repo.Created, 1)
	require.Equal(t, id, repo.Created[0].ID)
	require.Equal(t, "buy milk", repo.Created[0].Title)
}

func TestCreateTodoHandler_Handle_PublishesTodoCreatedEvent(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

	id, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "learn DDD"})

	require.NoError(t, err)
	require.Len(t, pub.Published, 1)

	event, ok := pub.Published[0].(todo.TodoCreated)
	require.True(t, ok)
	require.Equal(t, id, event.TodoID)
	require.Equal(t, "learn DDD", event.Title)
	require.NotZero(t, event.OccurredAt)
}

func TestCreateTodoHandler_Handle_RepoError(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{Err: errors.New("db down")}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

	_, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.EqualError(t, err, "db down")
	require.Empty(t, pub.Published)
}

func TestCreateTodoHandler_Handle_UoWError(t *testing.T) {
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Err: errors.New("tx failed")}, pub)

	_, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

	require.EqualError(t, err, "tx failed")
	require.Empty(t, pub.Published)
}

func TestCreateTodoHandler_Handle_EmptyTitle(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

	_, err := handler.Handle(context.Background(), CreateTodoCommand{Title: ""})

	require.ErrorIs(t, err, domain.ErrInvalidTitle)
	require.Empty(t, pub.Published)
}

func TestCreateTodoHandler_Handle_WhitespaceTitle(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	pub := &testhelpers.FakeEventPublisher{}
	handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

	_, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "   "})

	require.ErrorIs(t, err, domain.ErrInvalidTitle)
	require.Empty(t, pub.Published)
}
