package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/krzysztofkolcz/mymigrations/db"
	"github.com/stretchr/testify/require"
	slogctx "github.com/veqryn/slog-context"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// --- helpers ---

func nilLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertTableExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, schema, table string) {
	t.Helper()

	var exists bool
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
	)
	`
	err := pool.QueryRow(ctx, query, schema, table).Scan(&exists)
	require.NoError(t, err)
	require.Truef(t, exists, "table %s.%s should exist", schema, table)
}

func startPostgres(t *testing.T, ctx context.Context) (dsn string, pool *pgxpool.Pool, terminate func()) {
	t.Helper()

	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	dsn, err = pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	terminate = func() {
		pool.Close()
		_ = pg.Terminate(ctx)
	}

	return dsn, pool, terminate
}


func TestFullMigrationFlow(t *testing.T) {
	ctx := context.Background()

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	sqlDb, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create sqlDb", "err", err)
		os.Exit(1)
	}
	defer sqlDb.Close()

	// 1. public
	require.NoError(t, db.MigratePublic(ctx, dsn, logger))
	assertTableExists(t, ctx, pool, "public", "tenants")
	assertTableExists(t, ctx, pool, "public", "users")

	// 2. create tenant
	tenantID := uuid.New().String()
	schema := db.CreateSchemaName(tenantID)

	require.NoError(t, db.CreateTenant(ctx, pool, dsn, tenantID, nilLogger()))

	// 3. migrate tenant(s)
	require.NoError(t, db.MigrateAllTenants(ctx, pool, sqlDb, dsn, logger))

	// 4. verify one table from tenant migrations
	assertTableExists(t, ctx, pool, schema, "todos")

	// 5. version check
	var version int
	err = pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT version FROM "%s".schema_migrations`, schema),
	).Scan(&version)
	require.NoError(t, err)

	const latestVersion = 1
	require.Equal(t, latestVersion, version)
}

func TestTenantMigrations_AllTables(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	sqlDb, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create sqlDb", "err", err)
		os.Exit(1)
	}
	defer sqlDb.Close()

	require.NoError(t, db.MigratePublic(ctx, dsn, logger))

	tenantID := uuid.New().String()
	schema := "tenant_test"

	// create schema + wpis do tenants
	_, err = pool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema))
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO tenants (id, schema_name) VALUES ($1, $2)`,
		tenantID,
		schema,
	)
	require.NoError(t, err)

	require.NoError(t, db.MigrateAllTenants(ctx, pool, sqlDb, dsn, logger))

	// 🔴 dostosuj do swoich migracji
	expectedTables := []string{
		"todos",
	}

	for _, table := range expectedTables {
		assertTableExists(t, ctx, pool, schema, table)
	}
}

// --- TEST 3: idempotencja ---

func TestMigrations_Idempotent(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	sqlDb, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create sqlDb", "err", err)
		os.Exit(1)
	}
	defer sqlDb.Close()

	require.NoError(t, db.MigratePublic(ctx,dsn,logger))

	tenantID := uuid.New().String()

	require.NoError(t, db.CreateTenant(ctx, pool, dsn, tenantID, nilLogger()))

	// 🔁 uruchamiamy wielokrotnie
	require.NoError(t, db.MigrateAllTenants(ctx, pool, sqlDb, dsn, logger))
	require.NoError(t, db.MigrateAllTenants(ctx, pool, sqlDb, dsn, logger))
	require.NoError(t, db.MigrateAllTenants(ctx, pool, sqlDb, dsn, logger))

	// jeśli coś by się wysypało (duplikaty, konflikty), test padnie
}

// --- TEST 4: pełny scenariusz multi-tenant ---

func TestFullMigrationFlow_MultipleTenants(t *testing.T) {
	ctx := context.Background()

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))


	sqlDb, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create sqlDb", "err", err)
		os.Exit(1)
	}
	defer sqlDb.Close()

	require.NoError(t, db.MigratePublic(ctx, dsn, logger))

	tenantCount := 3
	schemas := make([]string, 0, tenantCount)

	for i := 0; i < tenantCount; i++ {
		tenantID := uuid.New().String()
		schema := db.CreateSchemaName(tenantID)
		schemas = append(schemas, schema)

		require.NoError(t, db.CreateTenant(ctx, pool, dsn, tenantID, nilLogger()))
	}

	require.NoError(t, db.MigrateAllTenants(ctx, pool, sqlDb, dsn, logger))

	for _, schema := range schemas {
		assertTableExists(t, ctx, pool, schema, "todos")
	}
}