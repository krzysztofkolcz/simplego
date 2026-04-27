package database

import (
	"context"
	"fmt"

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

func MigrateAllTenants(ctx context.Context, db *pgxpool.Pool, baseDSN string) error {
	rows, err := db.Query(ctx, `SELECT schema_name FROM tenants`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schema string
		rows.Scan(&schema)

		if err := MigrateTenant(baseDSN, schema); err != nil {
			return err
		}
	}
	return nil
}