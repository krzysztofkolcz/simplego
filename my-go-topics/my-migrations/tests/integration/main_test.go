package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/krzysztofkolcz/mymigrations/db"
)

var testDB *pgxpool.Pool
var testDSN string

func TestMain(m *testing.M) {
	fmt.Println("TestMain")
	ctx := context.Background()

	// container, err := postgres.Run(ctx,
	// 	"postgres:15",
	// 	// testcontainers.WithReuse(true), // szybciej - nie dziala.
	// )
	container, err := postgres.Run(ctx,
		"postgres:15",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	testDSN, err = container.ConnectionString(ctx, "sslmode=disable")
	fmt.Println(testDSN)
	if err != nil {
		log.Fatal(err)
	}

	testDB, err = pgxpool.New(ctx, testDSN)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := testDB.Ping(ctx); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// MIGRACJE
	if err := db.MigratePublic(testDSN); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	tenantID := uuid.New().String()
	if err := db.CreateTenant(ctx, testDB, testDSN, tenantID); err != nil {
		log.Fatal(err)
	}
	
	if err := db.MigrateAllTenants(ctx, testDB, testDSN); err != nil {
		log.Fatal(err)
	}
	code := m.Run()

	os.Exit(code)
}