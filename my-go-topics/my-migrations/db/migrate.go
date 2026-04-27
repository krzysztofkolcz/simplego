package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MigratePublic(dsn string) error {
	m, err := migrate.New(
		"file://db/migrations/public",
		dsn,
	)
	if err != nil {
		return err
	}

	return m.Up()
}

func CreateSchemaName(tenantID string ) string {
	safeID := "tenant_" + strings.ReplaceAll(tenantID, "-", "_")
	return safeID
}

func CreateTenant(ctx context.Context, pool *pgxpool.Pool, dsn, tenantID string) error {
	schema := CreateSchemaName(tenantID)

	// 1. create schema
	_, err := pool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema))
	if err != nil {
		return err
	}

	// 2. zapis do public.tenants
	_, err = pool.Exec(ctx,
		`INSERT INTO tenants (id, schema_name) VALUES ($1, $2)`,
		tenantID,
		schema,
	)
	if err != nil {
		return err
	}

	// 3. migracja
	return MigrateTenant(dsn, schema)
}

func MigrateTenant(dsn string, schema string) error {
	dsnWithSchema := fmt.Sprintf("%s&search_path=%s", dsn, schema)

	m, err := migrate.New(
		"file://db/migrations/tenant",
		dsnWithSchema,
	)
	if err != nil {
		return err
	}

	return m.Up()
}

func MigrateAllTenants(ctx context.Context, pool *pgxpool.Pool, baseDSN string) error {
	rows, err := pool.Query(ctx, `SELECT schema_name FROM tenants`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return err
		}

		if err := MigrateTenant(baseDSN, schema); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("tenant %s migration failed: %w", schema, err)
		}
	}

	return nil
}
