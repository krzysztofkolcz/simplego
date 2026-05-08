package db

import (
	"context"
	"fmt"
)

type Tenant struct {
	ID     string
	Schema string
}

func (t Tenant) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("tenant id is required")
	}

	if t.Schema == "" {
		return fmt.Errorf("tenant schema is required")
	}

	return nil
}

type tenantContextKey struct{}

func ContextWithTenant(
	ctx context.Context,
	tenant Tenant,
) context.Context {

	return context.WithValue(
		ctx,
		tenantContextKey{},
		tenant,
	)
}

func TenantFromContext(
	ctx context.Context,
) (Tenant, error) {

	tenant, ok := ctx.Value(
		tenantContextKey{},
	).(Tenant)

	if !ok {
		return Tenant{}, fmt.Errorf("tenant missing from context")
	}

	return tenant, nil
}