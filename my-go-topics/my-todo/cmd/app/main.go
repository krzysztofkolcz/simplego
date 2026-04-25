package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/app/commands"
	"github.com/C5383717/my-todo/internal/app/queries"
	"github.com/C5383717/my-todo/internal/config"
	"github.com/C5383717/my-todo/internal/constants"
	"github.com/C5383717/my-todo/internal/daemon"
	"github.com/C5383717/my-todo/internal/infrastructure/events"
	"github.com/C5383717/my-todo/internal/infrastructure/postgres"
	logger "github.com/C5383717/my-todo/internal/log"
	"github.com/C5383717/my-todo/internal/runner"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/oops"
)

var (
	gracefulShutdownSec     = flag.Int64("graceful-shutdown", 1, "graceful shutdown seconds")
	gracefulShutdownMessage = flag.String("graceful-shutdown-message", "Graceful shutdown in %d seconds",
		"graceful shutdown message")
)

// main is kept intentionally small — hard to test.
// All real work happens in run().
func main() {
	flag.Parse()

	exitCode := runner.RunFuncWithSignalHandling(run, runner.RunFlags{
		GracefulShutdownSec:     *gracefulShutdownSec,
		GracefulShutdownMessage: *gracefulShutdownMessage,
		Env:                     constants.APIName,
	})
	os.Exit(exitCode)
}

// run wires the entire dependency graph.
//
// Reading this function top to bottom reveals the full architecture:
//  1. Infrastructure: open DB pool
//  2. Infrastructure adapters: implement domain ports
//  3. Application: build command handlers (depend on domain ports)
//  4. Application: build query handlers (depend on domain ports)
//  5. Application: wire buses (map command/query types → handlers)
//  6. HTTP: create server (depends on buses only)
//
// Each layer depends only on interfaces from the layer below it.
// No layer imports a layer above it.
func run(ctx context.Context, cfg *config.Config) error {
	logger.InitLogger(*cfg)

	logger.Debug(ctx, "Starting the application", slog.Any("config", cfg))

	// ── 1. Infrastructure: database connection pool ─────────────────────────
	pool, err := pgxpool.New(ctx, cfg.Database.DSN)
	if err != nil {
		return oops.In("main").Wrapf(err, "opening database connection pool")
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return oops.In("main").Wrapf(err, "pinging database")
	}

	logger.Info(ctx, "Database connection established")

	// ── 2. Infrastructure adapters (implement domain PORTS) ─────────────────
	todoRepo := postgres.NewTodoRepository(pool)     // implements domain.TodoRepository
	eventPublisher := events.NewInMemoryEventPublisher() // implements domain.EventPublisher

	// ── 3. Application: command handlers ────────────────────────────────────
	// Each handler receives domain ports (interfaces), never concrete types.
	// This means handlers are testable with mock repositories.
	createHandler  := commands.NewCreateTodoHandler(todoRepo, eventPublisher)
	completeHandler := commands.NewCompleteTodoHandler(todoRepo, eventPublisher)
	updateHandler  := commands.NewUpdateTodoHandler(todoRepo, eventPublisher)
	deleteHandler  := commands.NewDeleteTodoHandler(todoRepo, eventPublisher)

	// ── 4. Application: query handlers ──────────────────────────────────────
	// Queries are read-only — they only need the repository, not the event publisher.
	getHandler  := queries.NewGetTodoByIDHandler(todoRepo)
	listHandler := queries.NewListTodosHandler(todoRepo)

	// ── 5. Application: wire command bus ────────────────────────────────────
	// Register maps command type → handler. Zero-value structs are used as
	// type keys — only reflect.TypeOf() is called on them.
	cmdBus := bus.NewInMemoryCommandBus()
	cmdBus.Register(commands.CreateTodoCommand{},   createHandler)
	cmdBus.Register(commands.CompleteTodoCommand{}, completeHandler)
	cmdBus.Register(commands.UpdateTodoCommand{},   updateHandler)
	cmdBus.Register(commands.DeleteTodoCommand{},   deleteHandler)

	// ── 6. Application: wire query bus ──────────────────────────────────────
	qryBus := bus.NewInMemoryQueryBus()
	qryBus.Register(queries.GetTodoByIDQuery{}, getHandler)
	qryBus.Register(queries.ListTodosQuery{},   listHandler)

	// ── 7. HTTP inbound adapter ──────────────────────────────────────────────
	s, err := daemon.NewTodoServer(ctx, cfg, cmdBus, qryBus)
	if err != nil {
		return oops.In("main").Wrapf(err, "creating api server")
	}

	if err := s.Start(ctx); err != nil {
		return oops.In("main").Wrapf(err, "starting api server")
	}

	logger.Info(ctx, "TODO API Server has started", slog.String("address", cfg.HTTP.Address))

	<-ctx.Done()

	if err := s.Close(ctx); err != nil {
		return oops.In("main").Wrapf(err, "closing server")
	}

	return nil
}
