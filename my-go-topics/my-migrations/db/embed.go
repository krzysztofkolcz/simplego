package db

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"embed"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

// d, err := iofs.New(migrationsFS, "migrations")
// if err != nil {
//     panic(err)
// }

// m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
// if err != nil {
//     panic(err)
// }

// err = m.Up()
// if err != nil && err != migrate.ErrNoChange {
//     panic(err)
// }

// 🔥 embed wszystkiego naraz

//go:embed migrations/*
var MigrationsFS embed.FS



func MigratePublicEmbed(dsn string) error {
	source, err := iofs.New(MigrationsFS, "migrations/public")
	if err != nil {
		return fmt.Errorf("iofs public: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("migrate public: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func MigrateTenantEmbed(dsn string, schema string) error {
	source, err := iofs.New(MigrationsFS, "migrations/tenant")
	if err != nil {
		return fmt.Errorf("iofs tenant: %w", err)
	}

	dsnWithSchema := fmt.Sprintf("%s&search_path=%s", dsn, schema)

	m, err := migrate.NewWithSourceInstance("iofs", source, dsnWithSchema)
	if err != nil {
		return fmt.Errorf("migrate tenant: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func MigrateAllTenantsEmbed(ctx context.Context, pool *pgxpool.Pool, dsn string) error {
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

		if err := MigrateTenantEmbed(dsn, schema); err != nil {
			return fmt.Errorf("tenant %s: %w", schema, err)
		}
	}

	return nil
}

func MigrateAllEbmed(ctx context.Context, pool *pgxpool.Pool, dsn string) error {
	if err := MigratePublicEmbed(dsn); err != nil {
		return err
	}

	return MigrateAllTenantsEmbed(ctx, pool, dsn)
}