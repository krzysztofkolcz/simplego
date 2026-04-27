package handler

import (
	"net/http"

	"github.com/krzysztofkolcz/my-sqlc/internal/database"
	"github.com/krzysztofkolcz/my-sqlc/internal/repository"
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := r.Header.Get("X-Tenant-ID")
	schema := "tenant_" + tenantID

	q, cleanup, err := database.WithTenant(ctx, dbPool, schema)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer cleanup()

	repo := repository.NewOrderRepo(q)

	// używasz repo normalnie
}