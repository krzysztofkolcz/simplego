package apierrors

import (
	"net/http"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
	"github.com/C5383717/my-todo/internal/domain"
)

// todoMapper maps domain errors to HTTP error responses.
// These are registered in the apiErrorMapper via slices.Concat in mapping.go.
var todoMapper = []APIErrors{
	{
		// ErrTodoNotFound is returned by the repository when a Todo does not exist.
		// Maps to 404 so callers can distinguish "not found" from "server error".
		Errors: []error{domain.ErrTodoNotFound},
		ExposedError: todoapi.DetailedError{
			Code:    "TODO_NOT_FOUND",
			Message: "Todo not found",
			Status:  http.StatusNotFound,
		},
	},
	{
		// ErrTodoAlreadyCompleted is enforced by the Todo aggregate.
		// Maps to 409 Conflict — the resource exists but is in the wrong state.
		Errors: []error{domain.ErrTodoAlreadyCompleted},
		ExposedError: todoapi.DetailedError{
			Code:    "TODO_ALREADY_COMPLETED",
			Message: "This todo is already completed",
			Status:  http.StatusConflict,
		},
	},
	{
		// ErrTitleRequired is enforced by the Todo aggregate constructor and Update.
		// Maps to 400 — the client sent invalid data.
		Errors: []error{domain.ErrTitleRequired},
		ExposedError: todoapi.DetailedError{
			Code:    "TODO_TITLE_REQUIRED",
			Message: "Todo title is required",
			Status:  http.StatusBadRequest,
		},
	},
}
