package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	slogctx "github.com/veqryn/slog-context"

	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
)

var ConnectionPool *pgxpool.Pool
var SqlDb *sql.DB
var TestDsn string
var TenantId string
var TenantSchema string
var ExpectedTables = []string{
	"todos",
	"outbox_events",
}

func TestMain(m *testing.M) {
	fmt.Println("TestMain")
	ctx := context.Background()

	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	container, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections"),
		),
		/*
			postgres.WithWaitStrategy(
				wait.ForAll(
					wait.ForListeningPort("5432/tcp"),
					wait.ForLog("database system is ready to accept connections"),
				),
			)
		*/
	)
	if err != nil {
		log.Fatal(err)
	}
	defer container.Terminate(ctx)

	TestDsn, err = container.ConnectionString(ctx, "sslmode=disable")
	fmt.Println(TestDsn)
	if err != nil {
		log.Fatal(err)
	}

	ConnectionPool, err = pgxpool.New(ctx, TestDsn)
	if err != nil {
		log.Fatal(err)
	}
	defer ConnectionPool.Close()

	SqlDb, err = sql.Open("pgx", TestDsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create sqlDb", "err", err)
		os.Exit(1)
	}
	defer SqlDb.Close()

	for i := 0; i < 10; i++ {
		if err := ConnectionPool.Ping(ctx); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// MIGRACJE
	if err := db.MigratePublic(ctx, TestDsn, logger); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	TenantId = uuid.New().String()
	TenantSchema = db.CreateSchemaName(TenantId)
	if err := db.CreateTenant(ctx, ConnectionPool, TenantId, logger); err != nil {
		log.Fatal(err)
	}

	if err := db.MigrateAllTenants(ctx, ConnectionPool, SqlDb, TestDsn, logger); err != nil {
		log.Fatal(err)
	}
	code := m.Run()

	os.Exit(code)
}
