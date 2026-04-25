package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/domain"
	"github.com/google/uuid"
)

// DeleteTodoCommand permanently removes a Todo by ID.
type DeleteTodoCommand struct {
	ID uuid.UUID
}

type DeleteTodoHandler struct {
	repo      domain.TodoRepository
	publisher domain.EventPublisher
}

func NewDeleteTodoHandler(repo domain.TodoRepository, publisher domain.EventPublisher) *DeleteTodoHandler {
	return &DeleteTodoHandler{repo: repo, publisher: publisher}
}

// Handle implements bus.CommandHandler.
// We verify existence before deletion so we can return 404 when appropriate.
func (h *DeleteTodoHandler) Handle(ctx context.Context, cmd bus.Command) error {
	c, ok := cmd.(DeleteTodoCommand)
	if !ok {
		return fmt.Errorf("DeleteTodoHandler: unexpected command type %T", cmd)
	}

	// Fail fast: verify the todo exists before issuing the delete.
	// This guarantees we return ErrTodoNotFound (→ 404) instead of silently
	// deleting nothing.
	if _, err := h.repo.GetByID(ctx, c.ID); err != nil {
		return fmt.Errorf("delete todo: verify: %w", err)
	}

	if err := h.repo.Delete(ctx, c.ID); err != nil {
		return fmt.Errorf("delete todo: persist: %w", err)
	}

	if err := h.publisher.Publish(ctx, domain.TodoDeleted{
		TodoID:     c.ID,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("delete todo: publish event: %w", err)
	}

	return nil
}
