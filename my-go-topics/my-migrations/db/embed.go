package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"embed"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

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

func MigrateTenantEmbed(ctx context.Context, dsn string, schema string, logger *slog.Logger) error {
	source, err := iofs.New(MigrationsFS, "migrations/tenant")
	if err != nil {
		return fmt.Errorf("iofs tenant: %w", err)
	}

	dsnWithSchema := fmt.Sprintf(
		"%s&search_path=%s&x-migrations-table=%s.schema_migrations", 
		dsn, 
		schema,
		schema,
	)
	logger.InfoContext(ctx, "MigrateTenantEmbed", "dnsWithSchema", dsnWithSchema)

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

func MigrateAllTenantsEmbed(ctx context.Context, pool *pgxpool.Pool, dsn string,logger *slog.Logger) error {
	rows, err := pool.Query(ctx, `SELECT schema_name FROM tenants`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			logger.ErrorContext(ctx, "MigratellTenantsEmbed", "Error schema rows.Scan", err.Error())
			return err
		}

		if err := MigrateTenantEmbedVersionTwo(ctx, dsn, schema, logger); err != nil {
			logger.ErrorContext(ctx, "MigratellTenantsEmbed", "MigrateTenantEmbed", err.Error())
			return fmt.Errorf("tenant %s: %w", schema, err)
		}
	}

	return nil
}

func MigrateAllEbmed(ctx context.Context, pool *pgxpool.Pool, dsn string,logger *slog.Logger) error {
	if err := MigratePublicEmbed(dsn); err != nil {
		return err
	}

	return MigrateAllTenantsEmbed(ctx, pool, dsn, logger)
}

func CreateTenantEmbed(ctx context.Context, pool *pgxpool.Pool, dsn, tenantID string, logger *slog.Logger) error {
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
	// return MigrateTenantEmbedVersionTwo(ctx, dsn, schema, logger)
	return nil
}