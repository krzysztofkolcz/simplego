package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/domain"
	"github.com/google/uuid"
)

// UpdateTodoCommand changes the mutable fields of a Todo.
// Only Title and Description can be updated — status changes are handled
// exclusively by CompleteTodoCommand. This separation is a CQRS principle:
// each operation has a single, focused command.
type UpdateTodoCommand struct {
	ID          uuid.UUID
	Title       string
	Description *string
}

type UpdateTodoHandler struct {
	repo      domain.TodoRepository
	publisher domain.EventPublisher
}

func NewUpdateTodoHandler(repo domain.TodoRepository, publisher domain.EventPublisher) *UpdateTodoHandler {
	return &UpdateTodoHandler{repo: repo, publisher: publisher}
}

// Handle implements bus.CommandHandler.
func (h *UpdateTodoHandler) Handle(ctx context.Context, cmd bus.Command) error {
	c, ok := cmd.(UpdateTodoCommand)
	if !ok {
		return fmt.Errorf("UpdateTodoHandler: unexpected command type %T", cmd)
	}

	todo, err := h.repo.GetByID(ctx, c.ID)
	if err != nil {
		return fmt.Errorf("update todo: load: %w", err)
	}

	// The aggregate validates the new title — returns domain.ErrTitleRequired if empty.
	if err := todo.Update(c.Title, c.Description); err != nil {
		return fmt.Errorf("update todo: %w", err)
	}

	if err := h.repo.Update(ctx, todo); err != nil {
		return fmt.Errorf("update todo: persist: %w", err)
	}

	if err := h.publisher.Publish(ctx, domain.TodoUpdated{
		TodoID:     todo.ID,
		NewTitle:   todo.Title,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("update todo: publish event: %w", err)
	}

	return nil
}
