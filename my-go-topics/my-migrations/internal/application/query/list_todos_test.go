package query

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/testhelpers"
	"github.com/stretchr/testify/require"
)

func TestListTodoHandler_Handle_CallsList(t *testing.T) {
	repo := &testhelpers.FakeTodoRepo{}
	handler := NewListTodosHandler(repo)
	query := ListTodosQuery{}

	result, err := handler.Handle(t.Context(), query)
	require.NoError(t, err)
	require.Equal(t, []TodoResult{}, result)

}

func TestListTodoHandler_Handle_ReturnsMappedTodos(t *testing.T) {
	id := uuid.New()
	repo := &testhelpers.FakeTodoRepo{
		Created: []todo.Todo{
			{
				ID:        id,
				Title:     "Buy milk",
				Completed: false,
			},
		},
	}
	handler := NewListTodosHandler(repo)
	query := ListTodosQuery{}

	result, err := handler.Handle(t.Context(), query)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, id, result[0].ID)
	require.Equal(t, "Buy milk", result[0].Title)

}
