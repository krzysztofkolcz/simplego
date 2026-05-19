package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	"github.com/stretchr/testify/require"
	slogctx "github.com/veqryn/slog-context"
)

func TestPublicMigrations(t *testing.T) {
	ctx := context.Background()

	// schema public
	assertSchemaExists(t, ctx, ConnectionPool, "public")
	// users
	assertTableExists(t, ctx, ConnectionPool, "public", "users")
	// tenants
	assertTableExists(t, ctx, ConnectionPool, "public", "tenants")

	assertTablesExists(t, ctx, ConnectionPool, TenantSchema, ExpectedTables)

	assertMigrationVersion(t, ctx, ConnectionPool, TenantSchema, 3)
}

func TestTenantMigrations(t *testing.T) {
	ctx := context.Background()

	tenantID := uuid.New().String()
	schema := db.CreateSchemaName(tenantID)

	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	// tworzymy tenant + migracje
	err := db.CreateTenant(ctx, ConnectionPool, tenantID, logger)
	if err != nil {
		t.Fatal(err)
	}

	err = db.MigrateAllTenants(ctx, ConnectionPool, SqlDb, TestDsn, logger)
	if err != nil {
		t.Fatal(err)
	}

	// schema istnieje?
	assertSchemaExists(t, ctx, ConnectionPool, schema)

	// tabela todos istnieje?
	assertTableExists(t, ctx, ConnectionPool, schema, "todos")
}

func TestTenantTableWorks(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	tenantID := uuid.New().String()
	schema := db.CreateSchemaName(tenantID)

	err := db.CreateTenant(ctx, ConnectionPool, tenantID, logger)
	if err != nil {
		t.Fatal(err)
	}

	err = db.MigrateAllTenants(ctx, ConnectionPool, SqlDb, TestDsn, logger)
	if err != nil {
		t.Fatal(err)
	}

	// używamy transakcji żeby SET LOCAL search_path nie zanieczyszczało puli połączeń
	tx, err := ConnectionPool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, fmt.Sprintf(`SET LOCAL search_path TO "%s"`, schema))
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO todos (id, title) VALUES (gen_random_uuid(), 'test')`)
	if err != nil {
		t.Fatal("insert failed -> migracja jest błędna")
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationsIdempotency(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	require.NoError(t, db.MigratePublic(ctx, TestDsn, logger))

	tenantID := uuid.New().String()
	require.NoError(t, db.CreateTenant(ctx, ConnectionPool, tenantID, logger))

	// 🔁 uruchamiamy wielokrotnie
	// jeśli coś by się wysypało (duplikaty, konflikty), test padnie
	require.NoError(t, db.MigrateAllTenants(ctx, ConnectionPool, SqlDb, TestDsn, logger))
	require.NoError(t, db.MigrateAllTenants(ctx, ConnectionPool, SqlDb, TestDsn, logger))
	require.NoError(t, db.MigrateAllTenants(ctx, ConnectionPool, SqlDb, TestDsn, logger))
}

func assertTablesExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, schema string, tables []string) {
	for _, table := range tables {
		assertTableExists(t, ctx, pool, schema, table)
	}
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

func assertSchemaExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, schema string) {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.schemata
			WHERE schema_name = $1
		)
	`

	err := pool.QueryRow(ctx, query, schema).Scan(&exists)
	require.NoError(t, err)
	require.Truef(t, exists, "schema %s should exist", schema)
}

func assertMigrationVersion(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	schema string,
	expectedVersion int,
) {
	var version int
	err := pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT version FROM "%s".schema_migrations`, schema),
	).Scan(&version)
	require.NoError(t, err)

	require.Equal(t, expectedVersion, version)
}
