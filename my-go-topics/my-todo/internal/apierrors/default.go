package apierrors

import (
	"net/http"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
)

const (
	ResourceNotFound = "RESOURCE_NOT_FOUND"
)

// defaultMapper provides fallback error mappings.
// It acts as a safety net — for example if a sqlc query returns pgx.ErrNoRows
// and the repository adapter did not wrap it as domain.ErrTodoNotFound.
var defaultMapper = []APIErrors{
	{
		Errors: []error{},
		ExposedError: todoapi.DetailedError{
			Code:    ResourceNotFound,
			Message: "Requested resource not found",
			Status:  http.StatusNotFound,
		},
	},
}
