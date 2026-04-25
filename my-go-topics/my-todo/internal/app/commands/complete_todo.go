package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/domain"
	"github.com/google/uuid"
)

// CompleteTodoCommand marks a single Todo as done.
// This is a separate command (not part of UpdateTodoCommand) because
// "completing" a todo is a distinct business operation with its own invariant:
// you cannot complete a todo that is already done (ErrTodoAlreadyCompleted).
// Keeping it separate makes the intent explicit and mirrors the separate HTTP endpoint.
type CompleteTodoCommand struct {
	ID uuid.UUID
}

type CompleteTodoHandler struct {
	repo      domain.TodoRepository
	publisher domain.EventPublisher
}

func NewCompleteTodoHandler(repo domain.TodoRepository, publisher domain.EventPublisher) *CompleteTodoHandler {
	return &CompleteTodoHandler{repo: repo, publisher: publisher}
}

// Handle implements bus.CommandHandler.
// Flow:
//  1. Load the aggregate from the repository
//  2. Call todo.Complete() — the aggregate enforces the invariant
//  3. Persist the updated aggregate
//  4. Publish TodoCompleted domain event
func (h *CompleteTodoHandler) Handle(ctx context.Context, cmd bus.Command) error {
	c, ok := cmd.(CompleteTodoCommand)
	if !ok {
		return fmt.Errorf("CompleteTodoHandler: unexpected command type %T", cmd)
	}

	todo, err := h.repo.GetByID(ctx, c.ID)
	if err != nil {
		// domain.ErrTodoNotFound — apierrors maps this to 404
		return fmt.Errorf("complete todo: load: %w", err)
	}

	// The aggregate enforces its own invariant here.
	// If todo.Status == StatusDone, Complete() returns domain.ErrTodoAlreadyCompleted
	// which apierrors maps to 409 Conflict.
	if err := todo.Complete(); err != nil {
		return fmt.Errorf("complete todo: %w", err)
	}

	if err := h.repo.Update(ctx, todo); err != nil {
		return fmt.Errorf("complete todo: persist: %w", err)
	}

	if err := h.publisher.Publish(ctx, domain.TodoCompleted{
		TodoID:     todo.ID,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("complete todo: publish event: %w", err)
	}

	return nil
}
