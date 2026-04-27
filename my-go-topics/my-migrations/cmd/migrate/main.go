package main

import (
	"context"
	"log"

	mig "github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krzysztofkolcz/mymigrations/db"
)

func main() {
	// dsn := os.Getenv("DB_URL")

	// err := db.MigratePublic(dsn)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	ctx := context.Background()

	dsn := "postgres://mymigrationsuser:mypassword@localhost:5432/mymigrationsdb?sslmode=disable"
	pool, _ := pgxpool.New(ctx, dsn)

	migrate(ctx, pool, dsn)
}

func migrate(ctx context.Context ,pool *pgxpool.Pool, dsn string){
	if err := db.MigratePublic(dsn); err != nil && err != mig.ErrNoChange {
		log.Fatal(err)
	}

	tenantID := uuid.New().String()
	if err := db.CreateTenant(ctx, pool, dsn, tenantID); err != nil {
		log.Fatal(err)
	}
	
	if err := db.MigrateAllTenants(ctx, pool, dsn); err != nil {
		log.Fatal(err)
	}
}