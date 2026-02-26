package apierrors

import (
	"errors"
	"net/http"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
)

const (
	TenantNotFound = "TENANT_NOT_FOUND"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
)

var highPrio = []APIErrors{
	{
		Errors: []error{ErrTenantNotFound},
		ExposedError: myhttpserver.DetailedError{
			Code:    TenantNotFound,
			Message: "Tenant does not exist",
			Status:  http.StatusNotFound,
		},
	},
	// {
	// 	Errors: []error{repo.ErrTenantNotFound},
	// 	ExposedError: cmkapi.DetailedError{
	// 		Code:    TenantNotFound,
	// 		Message: "Tenant does not exist",
	// 		Status:  http.StatusNotFound,
	// 	},
	// },
}
