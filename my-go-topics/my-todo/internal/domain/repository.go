package domain

import (
	"context"

	"github.com/google/uuid"
)

// TodoRepository is the repository PORT in Hexagonal Architecture terms.
//
// A PORT is an interface defined inside the domain (or application) layer that
// describes what the domain needs from the outside world — without specifying
// HOW it is implemented. The actual implementation (PostgreSQL, in-memory, etc.)
// lives in the infrastructure layer and is injected at startup.
//
// This inversion of dependencies is the key insight of Hexagonal Architecture:
// the domain layer does NOT import the infrastructure layer. Instead, the
// infrastructure layer imports the domain and implements its ports.
type TodoRepository interface {
	// Create persists a new Todo. The Todo has already been validated and
	// constructed by domain.NewTodo before reaching this method.
	Create(ctx context.Context, todo *Todo) error

	// GetByID retrieves a Todo by its UUID primary key.
	// Returns ErrTodoNotFound when no matching record exists.
	GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)

	// List returns all Todos ordered by creation time (newest first).
	List(ctx context.Context) ([]Todo, error)

	// Update persists the current state of a Todo aggregate.
	// The caller is responsible for mutating the aggregate (via Complete or
	// Update methods) before passing it here.
	Update(ctx context.Context, todo *Todo) error

	// Delete removes a Todo permanently.
	// Returns ErrTodoNotFound when no matching record exists.
	Delete(ctx context.Context, id uuid.UUID) error
}
