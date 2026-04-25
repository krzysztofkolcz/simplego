package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/domain"
	"github.com/google/uuid"
)

// GetTodoByIDQuery is the read-side request to fetch a single Todo.
type GetTodoByIDQuery struct {
	ID uuid.UUID
}

// TodoView is the read model (DTO) for a single Todo.
//
// This is a key CQRS concept: the read side returns DTOs, not domain aggregates.
// The aggregate (domain.Todo) has behaviour; the DTO is a flat data bag.
// Keeping them separate prevents the domain from being shaped by HTTP concerns.
//
// TodoView lives in the queries package because it is a read-side concern.
// The write side (commands) never references this type.
type TodoView struct {
	ID          uuid.UUID
	Title       string
	Description *string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GetTodoByIDHandler struct {
	repo domain.TodoRepository
}

func NewGetTodoByIDHandler(repo domain.TodoRepository) *GetTodoByIDHandler {
	return &GetTodoByIDHandler{repo: repo}
}

// Handle implements bus.QueryHandler.
// Returns (TodoView, nil) on success or (nil, error) on failure.
// The HTTP handler type-asserts the result to TodoView.
func (h *GetTodoByIDHandler) Handle(ctx context.Context, q bus.Query) (any, error) {
	query, ok := q.(GetTodoByIDQuery)
	if !ok {
		return nil, fmt.Errorf("GetTodoByIDHandler: unexpected query type %T", q)
	}

	todo, err := h.repo.GetByID(ctx, query.ID)
	if err != nil {
		return nil, fmt.Errorf("get todo by id: %w", err)
	}

	return toTodoView(todo), nil
}

func toTodoView(t *domain.Todo) TodoView {
	return TodoView{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
