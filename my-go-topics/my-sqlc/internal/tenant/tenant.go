package tenant

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTenant(ctx context.Context, db *pgxpool.Pool, tenantID string) (string, error) {
	schema := "tenant_" + tenantID

	_, err := db.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema))
	if err != nil {
		return "", err
	}

	return schema, nil
}