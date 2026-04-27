package tests

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		// testcontainers.WithReuseByName()
	)
	if err != nil {
		log.Fatal(err)
	}


	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	testDB, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}

	// przy reuse MUSISZ zadbać o stan DB
	if err := resetDatabase(testDB); err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	// cleanup
	testDB.Close()
	if err := container.Terminate(ctx); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func resetDatabase(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
	`)
	if err != nil {
		return err
	}

	// migracje
	return runMigrations(db)
}