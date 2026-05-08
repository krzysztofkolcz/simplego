package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{
		pool: pool,
	}
}

func (m *TxManager) WithinTransaction(
	ctx context.Context,
	tenantSchema string,
	fn func(q *sqlc.Queries) error,
) error {

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// IMPORTANT:
	// LOCAL means transaction-scoped only.
	if _, err := tx.Exec(
		ctx,
		"SET LOCAL search_path = "+pgx.Identifier{tenantSchema}.Sanitize(),
	); err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	q := sqlc.New(tx)

	if err := fn(q); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}