package queries

import (
	"context"
	"fmt"

	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/domain"
)

// ListTodosQuery requests all Todo items.
// It is intentionally empty — in a real application you would add filter,
// pagination, and sorting fields here.
type ListTodosQuery struct{}

type ListTodosHandler struct {
	repo domain.TodoRepository
}

func NewListTodosHandler(repo domain.TodoRepository) *ListTodosHandler {
	return &ListTodosHandler{repo: repo}
}

// Handle implements bus.QueryHandler.
// Returns ([]TodoView, nil). The HTTP handler type-asserts to []TodoView.
func (h *ListTodosHandler) Handle(ctx context.Context, q bus.Query) (any, error) {
	if _, ok := q.(ListTodosQuery); !ok {
		return nil, fmt.Errorf("ListTodosHandler: unexpected query type %T", q)
	}

	todos, err := h.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}

	views := make([]TodoView, 0, len(todos))
	for _, t := range todos {
		t := t // capture loop variable
		views = append(views, toTodoView(&t))
	}

	return views, nil
}
