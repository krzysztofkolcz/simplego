package domain

import "errors"

// Domain errors are sentinel values defined at the domain layer.
// They travel up through the application layer into the HTTP adapter,
// where apierrors.TransformToAPIError maps them to HTTP status codes.
// No HTTP status codes or HTTP concepts live here.

var (
	// ErrTodoNotFound is returned by the repository when a Todo with the
	// requested ID does not exist in the data store.
	ErrTodoNotFound = errors.New("todo not found")

	// ErrTodoAlreadyCompleted is returned by Todo.Complete() when the
	// aggregate is already in the "done" state — completing it again is a
	// business rule violation.
	ErrTodoAlreadyCompleted = errors.New("todo already completed")

	// ErrTitleRequired is returned by NewTodo and Todo.Update() when an
	// empty title is provided. A Todo must always have a non-empty title.
	ErrTitleRequired = errors.New("todo title is required")
)
