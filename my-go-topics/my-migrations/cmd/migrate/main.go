package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	mig "github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krzysztofkolcz/mymigrations/db"
	slogctx "github.com/veqryn/slog-context"
)

func main() {
	mode := "serve"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	ctx := context.Background()
	dsn := buildDSNFromEnv()


	handler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	)
	logger := slog.New(handler)

	logger.InfoContext(ctx, "some info")

	// dsn := "postgres://mymigrationsuser:mypassword@localhost:5432/mymigrationsdb?sslmode=disable"
	pool, _ := pgxpool.New(ctx, dsn)
	slog.Info("dsn", "", dsn)

	switch mode {
	case "migrate":
		migrate(ctx, pool, dsn, logger)
	case "serve":
		runServer()
	default:
		log.Fatalf("unknown mode: %s", mode)
	}

	
}

func migrate(ctx context.Context ,pool *pgxpool.Pool, dsn string, logger *slog.Logger){
	err := db.ValidateMigrations(db.MigrationsFS, "migrations/tenant")
	if err != nil {
		log.Fatal("invalid tenant migrations:", err)
	}

	err = db.ValidateMigrations(db.MigrationsFS, "migrations/public")
	if err != nil {
		log.Fatal("invalid public migrations:", err)
	}

	if err := db.MigratePublicEmbed(dsn); err != nil && err != mig.ErrNoChange {
		log.Fatal(err)
	}

	// tenantID := uuid.New().String()
	// if err := db.CreateTenantEmbed(ctx, pool, dsn, tenantID, logger); err != nil {
	// 	log.Fatal(err)
	// }
	
	if err := db.MigrateAllTenantsEmbed(ctx, pool, dsn, logger); err != nil {
		log.Fatal(err)
	}
}

func buildDSNFromEnv()string{
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbname := os.Getenv("DB_NAME")
	dbuser := os.Getenv("DB_USER")
	dbpass := os.Getenv("DB_PASS")
	dsn := "postgres://"+dbuser+":"+dbpass+"@"+dbHost+":"+dbPort+"/"+dbname+"?sslmode=disable"
	return dsn

}


func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("APP_NAME")
	if name == "" {
		name = "go-migrations"
	}
	dbname := os.Getenv("DB_NAME")
	dbpass := os.Getenv("DB_PASS")
	fmt.Fprintf(w, "X Hello from %s! DB_NAME: %s, DB_PASS: %s", name, dbname, dbpass)
}

func runServer() {
	http.HandleFunc("/", handler)
	port := ":8080"
	http.ListenAndServe(port, nil)
}
