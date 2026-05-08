package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(
	ctx context.Context,
	cfg Config,
	logger *slog.Logger,
) (*pgxpool.Pool, error) {

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Called before a connection is handed to the app
	poolConfig.PrepareConn = func(
		ctx context.Context,
		conn *pgx.Conn,
	) (bool, error) {

		// optional validation here

		return true, nil
	}

	// Called after connection is released back to pool
	poolConfig.AfterRelease = func(conn *pgx.Conn) bool {
		return true
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := ping(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	logger.Info(
		"postgres pool initialized",
		slog.Int64("max_conns", int64(cfg.MaxConns)),
		slog.Int64("min_conns", int64(cfg.MinConns)),
	)

	return pool, nil
}

func ping(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
