package query

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/tenant"
)

type GetTenantQuery struct {
	ID uuid.UUID
}

type GetTenantResult struct {
	ID               uuid.UUID
	SchemaName       string
	MigrationStatus  string
	MigrationError   string
	MigrationUpdated time.Time
	CreatedAt        time.Time
}

type GetTenantHandler struct {
	repo tenant.Repository
}

func NewGetTenantHandler(repo tenant.Repository) *GetTenantHandler {
	return &GetTenantHandler{repo: repo}
}

func (h *GetTenantHandler) Handle(
	ctx context.Context,
	q GetTenantQuery,
) (*GetTenantResult, error) {

	t, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &GetTenantResult{
		ID:               t.ID,
		SchemaName:       t.SchemaName,
		MigrationStatus:  t.MigrationStatus,
		MigrationError:   t.MigrationError,
		MigrationUpdated: t.MigrationUpdated,
		CreatedAt:        t.CreatedAt,
	}, nil
}
