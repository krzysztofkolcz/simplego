package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/mymigrations/db"
	slogctx "github.com/veqryn/slog-context"
)

func TestPublicMigrations(t *testing.T) {
	ctx := context.Background()

	// schema public
	ok, err := schemaExists(ctx, testDB, "public")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("public schema does not exist")
	}

	// users
	ok, err = tableExists(ctx, testDB, "public", "users")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("users table missing")
	}

	// tenants
	ok, err = tableExists(ctx, testDB, "public", "tenants")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("tenants table missing")
	}
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
	err := db.CreateTenant(ctx, testDB, testDSN, tenantID, logger)
	if err != nil {
		t.Fatal(err)
	}

	// schema istnieje?
	ok, err := schemaExists(ctx, testDB, schema)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("schema %s not created", schema)
	}

	// tabela todos istnieje?
	ok, err = tableExists(ctx, testDB, schema, "todos")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("todos table missing in %s", schema)
	}
}

func TestTenantTableWorks(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	))

	tenantID := uuid.New().String()
	schema := db.CreateSchemaName(tenantID)

	err := db.CreateTenant(ctx, testDB, testDSN, tenantID, logger)
	if err != nil {
		t.Fatal(err)
	}

	// ustawiamy search_path
	_, err = testDB.Exec(ctx,
		fmt.Sprintf(`SET search_path TO %s`, schema),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testDB.Exec(ctx, `
		INSERT INTO todos (id, title) VALUES (gen_random_uuid(), 'test')
	`)
	if err != nil {
		t.Fatal("insert failed -> migracja jest błędna")
	}
}