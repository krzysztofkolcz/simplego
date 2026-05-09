package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	slogctx "github.com/veqryn/slog-context"
)

func main() {
	mode := "serve"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	ctx := context.Background()
	dsn := buildDSNFromEnv()

	slogHandler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	)
	logger := slog.New(slogHandler)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	sqlDb, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create sqlDb", "err", err)
		os.Exit(1)
	}
	defer sqlDb.Close()

	switch mode {
	case "migrate":
		if err := runMigrate(ctx, pool, sqlDb, dsn, logger); err != nil {
			logger.ErrorContext(ctx, "migration failed", "err", err)
			os.Exit(1)
		}
	case "serve":
		runServer(logger)
	case "retry-failed":
		err = retryFailed(ctx, pool, sqlDb, dsn, logger)
		if err != nil {
			logger.ErrorContext(ctx, "retry failed", "err", err)
		}
	default:
		logger.ErrorContext(ctx, "unknown mode", "mode", mode)
		os.Exit(1)
	}
}

func runMigrate(ctx context.Context, pool *pgxpool.Pool,sqlDb *sql.DB, dsn string, logger *slog.Logger) error {
	if err := db.ValidateMigrations(db.MigrationsFS, "migrations/tenant"); err != nil {
		return fmt.Errorf("invalid tenant migrations: %w", err)
	}

	if err := db.ValidateMigrations(db.MigrationsFS, "migrations/public"); err != nil {
		return fmt.Errorf("invalid public migrations: %w", err)
	}

	if err := db.MigratePublic(ctx, dsn, logger); err != nil {
		return err
	}

	tenantID := uuid.New().String()
	 db.CreateTenant(ctx, pool, tenantID, logger)

	return db.MigrateAllTenants(ctx, pool, sqlDb, dsn, logger)
}

func retryFailed(ctx context.Context, pool *pgxpool.Pool, sqlDb *sql.DB, dsn string, logger *slog.Logger) error {
	rows, err := pool.Query(ctx, `
		SELECT id, schema_name
		FROM tenants
		WHERE migration_status = 'failed'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, schema string
		rows.Scan(&id, &schema)
		err = db.MigrateSingleTenantWrapper(ctx, sqlDb, id, schema, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildDSNFromEnv() string {
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASS")),
		Host:     net.JoinHostPort(os.Getenv("DB_HOST"), dbPort),
		Path:     "/" + os.Getenv("DB_NAME"),
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("APP_NAME")
	if name == "" {
		name = "go-migrations"
	}
	fmt.Fprintf(w, "Hello from %s!", name)
}

func runServer(logger *slog.Logger) {
	http.HandleFunc("/", httpHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
