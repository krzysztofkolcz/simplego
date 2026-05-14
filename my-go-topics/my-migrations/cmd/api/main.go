package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krzysztofkolcz/mymigrations/internal/http/handler"
	"github.com/krzysztofkolcz/mymigrations/internal/http/router"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
	infraevent "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/event"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/usecase"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/worker"
	slogctx "github.com/veqryn/slog-context"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 30 * time.Second
)

func main() {
	// Logger
	slogHandler := slogctx.NewHandler(slog.NewJSONHandler(os.Stdout, nil), nil)
	logger := slog.New(slogHandler)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	// Database pool
	pool, err := db.NewPool(ctx, db.Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		MaxConns:          20,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Wiring
	txManager := db.NewTxManager(pool)
	commandQ := commanddb.New(pool)
	queryQ := querydb.New(pool)
	eventPublisher := infraevent.NewOutboxPublisher()

	outboxWorker := worker.NewOutboxWorker(txManager, commandQ, 5*time.Second)
	go outboxWorker.Run(ctx)

	srv := handler.NewServer(
		usecase.NewCreateTodoUseCase(txManager, eventPublisher),
		usecase.NewCompleteTodoUseCase(txManager, eventPublisher),
		usecase.NewDeleteTodoUseCase(txManager, eventPublisher),
		usecase.NewGetTodoUseCase(txManager),
		usecase.NewListTodosUseCase(txManager),
		usecase.NewCreateTenantUseCase(txManager),
		usecase.NewGetTenantUseCase(commandQ, queryQ),
		usecase.NewCreateUserUseCase(txManager),
		usecase.NewGetUserUseCase(commandQ, queryQ),
	)
	httpHandler := router.New(srv)

	// HTTP server
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           httpHandler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	// Start
	go func() {
		logger.Info("server starting", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Wait for signal
	<-ctx.Done()
	stop()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	logger.Info("shutdown complete")
}
