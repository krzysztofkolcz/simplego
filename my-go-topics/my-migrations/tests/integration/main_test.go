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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	slogctx "github.com/veqryn/slog-context"

	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
)

var testDB *pgxpool.Pool
var testDSN string

func TestMain(m *testing.M) {
	fmt.Println("TestMain")
	ctx := context.Background()

	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	container, err := postgres.Run(ctx,
		"postgres:15",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	fmt.Println(dsn)
	if err != nil {
		log.Fatal(err)
	}

	testDB, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	sqlDb, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create sqlDb", "err", err)
		os.Exit(1)
	}
	defer sqlDb.Close()


	for i := 0; i < 10; i++ {
		if err := testDB.Ping(ctx); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// MIGRACJE
	if err := db.MigratePublic(ctx, dsn, logger); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	tenantID := uuid.New().String()
	if err := db.CreateTenant(ctx, testDB, testDSN, tenantID, logger); err != nil {
		log.Fatal(err)
	}
	
	if err := db.MigrateAllTenants(ctx, testDB, sqlDb, testDSN, logger); err != nil {
		log.Fatal(err)
	}
	code := m.Run()

	os.Exit(code)
}