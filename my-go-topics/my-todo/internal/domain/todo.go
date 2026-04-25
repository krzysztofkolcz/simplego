// Package domain contains the core business logic for the TODO application.
// This package has zero imports from any other internal package — it depends
// only on the Go standard library and github.com/google/uuid.
// This enforces the Hexagonal Architecture rule: the domain is the innermost
// layer and knows nothing about HTTP, databases, or configuration.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of a Todo item.
// It is a value object — two Status values with the same string are equal.
type Status string

const (
	StatusPending Status = "pending"
	StatusDone    Status = "done"
)

// Todo is the aggregate root of the TODO bounded context.
// All business rules that govern a Todo item live here.
// The aggregate enforces its own invariants — no external code can put
// a Todo into an invalid state.
type Todo struct {
	ID          uuid.UUID
	Title       string
	Description *string // optional
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTodo is the only constructor for a Todo.
// It validates inputs and sets the initial state.
// The caller (command handler) generates the ID so it can be returned in the
// HTTP response without a round-trip query — a deliberate CQRS design choice.
func NewTodo(id uuid.UUID, title string, description *string) (*Todo, error) {
	if title == "" {
		return nil, ErrTitleRequired
	}

	now := time.Now().UTC()

	return &Todo{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Complete transitions the Todo from pending to done.
// It is an error to complete a Todo that is already done — the aggregate
// enforces this invariant so no caller can accidentally do it.
func (t *Todo) Complete() error {
	if t.Status == StatusDone {
		return ErrTodoAlreadyCompleted
	}

	t.Status = StatusDone
	t.UpdatedAt = time.Now().UTC()

	return nil
}

// Update changes the mutable fields of a Todo.
// Only title and description can be updated — ID, status, and timestamps
// are managed by the aggregate itself.
func (t *Todo) Update(title string, description *string) error {
	if title == "" {
		return ErrTitleRequired
	}

	t.Title = title
	t.Description = description
	t.UpdatedAt = time.Now().UTC()

	return nil
}

// Reconstitute rebuilds a Todo from persisted data (e.g., a database row).
// Unlike NewTodo, it does not set CreatedAt/UpdatedAt and accepts any Status.
// Use only in repository adapters — never call this from business logic.
func Reconstitute(
	id uuid.UUID,
	title string,
	description *string,
	status Status,
	createdAt time.Time,
	updatedAt time.Time,
) *Todo {
	return &Todo{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      status,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// EnsureValidStatus returns an error if s is not a recognised Status value.
// Used by the repository adapter when mapping DB rows to domain objects.
func EnsureValidStatus(s string) (Status, error) {
	switch Status(s) {
	case StatusPending, StatusDone:
		return Status(s), nil
	default:
		return "", errors.New("unknown todo status: " + s)
	}
}
