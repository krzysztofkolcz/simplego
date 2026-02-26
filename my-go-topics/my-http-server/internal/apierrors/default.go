package apierrors

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
)

const (
	ResourceNotFound = "RESOURCE_NOT_FOUND"
	UniqueError      = "UNIQUE_ERROR"
	BadRequest       = "BAD_REQUEST"
	GetResource      = "GET_RESOURCE"
)

var (
	ErrActionRequireWorkflow = errors.New("action requires a workflow")
	ErrUnknownProperty       = errors.New("unknown property")
	ErrBadOdataFilter        = errors.New("bad odata filter")
)

var defaultMapper = []APIErrors{
	{
		Errors: []error{sql.ErrNoRows},
		ExposedError: myhttpserver.DetailedError{
			Code:    ResourceNotFound,
			Message: "Requested resource not found",
			Status:  http.StatusNotFound,
		},
	},
	{
		Errors: []error{ErrBadOdataFilter},
		ExposedError: myhttpserver.DetailedError{
			Code:    BadRequest,
			Message: "Bad Odata filter provided",
			Status:  http.StatusBadRequest,
		},
	},
	{
		Errors: []error{ErrUnknownProperty},
		ExposedError: myhttpserver.DetailedError{
			Code:    "UNKNOWN_PROPERTY",
			Message: "Unknown property",
			Status:  http.StatusBadRequest,
		},
	},
	// {
	// 	Errors: []error{repo.ErrUniqueConstraint},
	// 	ExposedError: myhttpserver.DetailedError{
	// 		Code:    UniqueError,
	// 		Message: "Resource with such ID already exists",
	// 		Status:  http.StatusConflict,
	// 	},
	// },
}
