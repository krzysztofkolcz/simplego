package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func MigrateTenantEmbedVersionTwo(
	ctx context.Context,
	dsn string,
	schema string,
	logger *slog.Logger,
) error {

	// 1. Source (embed)
	source, err := iofs.New(MigrationsFS, "migrations/tenant")
	if err != nil {
		return fmt.Errorf("iofs tenant: %w", err)
	}

	files, err := fs.ReadDir(MigrationsFS, "migrations/tenant")
	if err != nil {
		logger.Error("read dir error", "err", err)
		return err
	}

	for _, f := range files {
		logger.Info("migration file", "name", f.Name())
	}

	// 2. sql.DB (WAŻNE: używamy stdlib drivera pgx)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer db.Close()

	// 3. ustaw search_path dla tego połączenia
	if _, err := db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", schema)); err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	// 4. driver migrate z tabelą per schema
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations", // nazwa tabeli
		SchemaName:      schema,              // 🔥 KLUCZOWE
	})
	if err != nil {
		return fmt.Errorf("postgres instance: %w", err)
	}

	// 5. migrate instance
	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migrate tenant: %w", err)
	}

	// 6. run
	logger.InfoContext(ctx, "BEFORE migrate", "schema", schema)
	err = m.Up()
	logger.InfoContext(ctx, "AFTER migrate", "schema", schema, "err", err)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	logger.InfoContext(ctx, "tenant migrated", "schema", schema)

	return nil
}