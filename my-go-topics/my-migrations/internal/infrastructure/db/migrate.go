package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*
var MigrationsFS embed.FS

func CreateSchemaName(tenantID string) string {
	return "tenant_" + strings.ReplaceAll(tenantID, "-", "_")
}

func MigratePublic(ctx context.Context, dsn string, logger *slog.Logger) error {
	logger.InfoContext(ctx, "migrate public")
	source, err := iofs.New(MigrationsFS, "migrations/public")
	if err != nil {
		logger.ErrorContext(ctx, "migrate public", "iofs public error", err.Error())
		return fmt.Errorf("iofs public: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		logger.ErrorContext(ctx, "migrate public", "error", err.Error())
		return fmt.Errorf("migrate public: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.ErrorContext(ctx, "migrate public", "m.Up error", err.Error())
		return err
	}
	logger.InfoContext(ctx, "migrate public success")
	return nil
}

func MigrateAllTenants(ctx context.Context, pool *pgxpool.Pool, sqlDb *sql.DB, dsn string, logger *slog.Logger) error {
	var failed []string
	rows, err := pool.Query(ctx, `SELECT id, schema_name FROM tenants`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, schema string
		if err := rows.Scan(&id, &schema); err != nil {
			logger.ErrorContext(ctx, "scan failed", "schema", schema, "err", err)
			return err
		}

		if err := MigrateSingleTenantWrapper(ctx, sqlDb, dsn, id, schema, logger); err != nil {
			logger.ErrorContext(ctx, "migrate tenant failed", "schema", schema, "err", err)
			failed = append(failed, schema)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration: %w", err)
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed tenants: %v", failed)
	}

	return nil
}

func CreateTenant(ctx context.Context, pool *pgxpool.Pool, dsn, tenantID string, logger *slog.Logger) error {
	schema := CreateSchemaName(tenantID)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO tenants (id, schema_name) VALUES ($1, $2)`,
		tenantID,
		schema,
	); err != nil {
		return fmt.Errorf("insert tenant: %w", err)
	}

	return tx.Commit(ctx)
}

func MigrateTenant(
	ctx context.Context,
	sqlDb *sql.DB,
	schema string,
	logger *slog.Logger,
) error {
	source, err := iofs.New(MigrationsFS, "migrations/tenant")
	if err != nil {
		return fmt.Errorf("iofs tenant: %w", err)
	}

	/*
		Claude.ai - https://claude.ai/chat/edae26e3-b9a2-4ec1-b339-21ef4db45606
		sqlDb.ExecContext(ctx, "SET search_path TO "+quotedSchema)
		sql.DB to pula połączeń — SET search_path ustawia stan na konkretnym połączeniu,
		ale kolejne zapytanie może trafić na inne.
		Przy współbieżnym wykonaniu (nawet sekwencyjnym z powrotem połączenia do puli)
		może dojść do migracji w złym schemacie.
		Rozwiązanie: użyj search_path w DSN-ie lub dedykowanego połączenia (sqlDb.Conn(ctx)):
	*/
	conn, err := sqlDb.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer func() {
		// reset search_path before returning to pool to avoid contaminating other queries
		_, _ = conn.ExecContext(ctx, "SET search_path TO public")
		conn.Close()
	}()

	quotedSchema := `"` + strings.ReplaceAll(schema, `"`, `""`) + `"`
	if _, err := conn.ExecContext(ctx, "SET search_path TO "+quotedSchema); err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	driver, err := postgres.WithConnection(ctx, conn, &postgres.Config{
		MigrationsTable: "schema_migrations",
		SchemaName:      schema,
	})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate tenant: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	logger.InfoContext(ctx, "tenant migrated", "schema", schema)
	return nil
}

func MigrateSingleTenantWrapper(
	ctx context.Context,
	sqlDb *sql.DB,
	dsn, tenantID, schema string,
	logger *slog.Logger,
) error {

	setStatus := func(status, errMsg string) {
		_, err := sqlDb.ExecContext(ctx, `
			UPDATE public.tenants
			SET migration_status = $1,
			    migration_error = $2,
			    migration_updated_at = NOW()
			WHERE id = $3
		`, status, errMsg, tenantID)

		if err != nil {
			logger.Error("failed to update tenant status",
				"tenantID", tenantID,
				"err", err,
			)
		}
	}

	setStatus("running", "")

	logger.Info("migrating tenant", "schema", schema)

	// retry
	var err error
	for i := 0; i < 3; i++ {
		err = MigrateTenant(ctx, sqlDb, schema, logger)
		if err == nil {
			break
		}

		logger.Warn("retry tenant",
			"schema", schema,
			"attempt", i+1,
			"err", err,
		)

		select {
		case <-time.After(time.Duration(1<<i) * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if err != nil {
		setStatus("failed", err.Error())

		logger.Error("tenant failed",
			"schema", schema,
			"err", err,
		)

		return err
	}

	setStatus("success", "")

	logger.Info("tenant success", "schema", schema)

	return nil
}
