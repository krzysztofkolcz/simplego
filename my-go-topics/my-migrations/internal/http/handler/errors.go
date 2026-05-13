package handler

import (
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
)

// tenantSchema converts a tenant UUID (from X-Tenant-ID header)
// to the PostgreSQL schema name for that tenant.
func tenantSchema(id httpapi.XTenantID) string {
	return db.CreateSchemaName(id.String())
}

func internalError(err error) httpapi.N500JSONResponse {
	return httpapi.N500JSONResponse{
		Error: httpapi.DetailedError{
			Status:  500,
			Code:    "INTERNAL_SERVER_ERROR",
			Message: err.Error(),
		},
	}
}

func notFoundError() httpapi.N404JSONResponse {
	return httpapi.N404JSONResponse{
		Error: httpapi.DetailedError{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Resource not found",
		},
	}
}

func conflictError() httpapi.N409JSONResponse {
	return httpapi.N409JSONResponse{
		Error: httpapi.DetailedError{
			Status:  409,
			Code:    "CONFLICT",
			Message: "Resource already exists",
		},
	}
}
