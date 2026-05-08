package api

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	slogctx "github.com/veqryn/slog-context"
)

func main() {
	 ctx := context.Background()
	slogHandler := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		nil,
	)
	logger := slog.New(slogHandler)
	pool, err := db.NewPool(
		ctx,
		db.Config{
			DatabaseURL: os.Getenv("DATABASE_URL"),

			MaxConns: 20,
			MinConns: 5,

			MaxConnLifetime: time.Hour,
			MaxConnIdleTime: 30 * time.Minute,

			HealthCheckPeriod: time.Minute,
		},
		logger,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()
}