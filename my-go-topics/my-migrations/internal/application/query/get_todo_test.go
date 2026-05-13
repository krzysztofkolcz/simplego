package query

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
)

func TestGetTodoHandler_Handle_ReturnsNotFound(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewGetTodoHandler(repo)

	_, err := handler.Handle(t.Context(), GetTodoQuery{ID: uuid.New()})

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetTodoHandler_Handle_ReturnsMappedTodo(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().Truncate(time.Second)
	repo := &testhelpers.FakeTodoRepo{
		Created: []todo.Todo{
			{ID: id, Title: "learn DDD", Completed: true, CreatedAt: createdAt},
		},
	}
	handler := NewGetTodoHandler(repo)

	result, err := handler.Handle(t.Context(), GetTodoQuery{ID: id})

	require.NoError(t, err)
	require.Equal(t, id, result.ID)
	require.Equal(t, "learn DDD", result.Title)
	require.True(t, result.Completed)
	require.Equal(t, createdAt, result.CreatedAt)
}
