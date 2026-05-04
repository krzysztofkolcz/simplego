package db_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/krzysztofkolcz/mymigrations/db"
	"github.com/stretchr/testify/require"

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

// --- TEST 1: Full flow ---

func TestFullMigrationFlow(t *testing.T) {
	ctx := context.Background()

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	// 1. public
	require.NoError(t, db.MigratePublicEmbed(dsn))

	// 2. create tenant
	tenantID := uuid.New().String()
	schema := "tenant_" + tenantID

	require.NoError(t, db.CreateTenantEmbed(ctx, pool, dsn, tenantID, nilLogger()))

	// 3. migrate tenant(s)
	require.NoError(t, db.MigrateAllTenantsEmbed(ctx, pool, dsn, nilLogger()))

	// 4. verify one table from tenant migrations
	assertTableExists(t, ctx, pool, schema, "test_123") // dostosuj nazwę

	// 5. version check
	var version int
	err := pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT version FROM "%s".schema_migrations`, schema),
	).Scan(&version)
	require.NoError(t, err)

	const latestVersion = 6 // ← dostosuj
	require.Equal(t, latestVersion, version)
}

// --- TEST 2: wszystkie tabele tenant ---

func TestTenantMigrations_AllTables(t *testing.T) {
	ctx := context.Background()

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	require.NoError(t, db.MigratePublicEmbed(dsn))

	tenantID := "1"
	schema := "tenant_test"

	// create schema + wpis do tenants
	_, err := pool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema))
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO tenants (id, schema_name) VALUES ($1, $2)`,
		tenantID,
		schema,
	)
	require.NoError(t, err)

	require.NoError(t, db.MigrateAllTenantsEmbed(ctx, pool, dsn, nilLogger()))

	// 🔴 dostosuj do swoich migracji
	expectedTables := []string{
		"test_123",
		// "todos",
		// "orders",
	}

	for _, table := range expectedTables {
		assertTableExists(t, ctx, pool, schema, table)
	}
}

// --- TEST 3: idempotencja ---

func TestMigrations_Idempotent(t *testing.T) {
	ctx := context.Background()

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	require.NoError(t, db.MigratePublicEmbed(dsn))

	tenantID := uuid.New().String()

	require.NoError(t, db.CreateTenantEmbed(ctx, pool, dsn, tenantID, nilLogger()))

	// 🔁 uruchamiamy wielokrotnie
	require.NoError(t, db.MigrateAllTenantsEmbed(ctx, pool, dsn, nilLogger()))
	require.NoError(t, db.MigrateAllTenantsEmbed(ctx, pool, dsn, nilLogger()))
	require.NoError(t, db.MigrateAllTenantsEmbed(ctx, pool, dsn, nilLogger()))

	// jeśli coś by się wysypało (duplikaty, konflikty), test padnie
}

// --- TEST 4: pełny scenariusz multi-tenant ---

func TestFullMigrationFlow_MultipleTenants(t *testing.T) {
	ctx := context.Background()

	dsn, pool, terminate := startPostgres(t, ctx)
	defer terminate()

	require.NoError(t, db.MigratePublicEmbed(dsn))

	tenantCount := 3
	schemas := make([]string, 0, tenantCount)

	for i := 0; i < tenantCount; i++ {
		tenantID := uuid.New().String()
		schema := "tenant_" + tenantID
		schemas = append(schemas, schema)

		require.NoError(t, db.CreateTenantEmbed(ctx, pool, dsn, tenantID, nilLogger()))
	}

	require.NoError(t, db.MigrateAllTenantsEmbed(ctx, pool, dsn, nilLogger()))

	for _, schema := range schemas {
		assertTableExists(t, ctx, pool, schema, "test_123")
	}
}