package db

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

var schemaRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(
	pool *pgxpool.Pool,
) *TxManager {

	return &TxManager{
		pool: pool,
	}
}

func (m *TxManager) WithinTransaction(
	ctx context.Context,
	tenantSchema string,
	fn func(q *commanddb.Queries) error,
) error {

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {

		return fmt.Errorf(
			"begin tx: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := setTenantSchema(ctx, tx, tenantSchema); err != nil {
		return err
	}

	q := commanddb.New(tx)

	if err := fn(q); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"commit tx: %w",
			err,
		)
	}

	return nil
}

func setTenantSchema(
	ctx context.Context,
	tx pgx.Tx,
	schema string,
) error {

	query := fmt.Sprintf(`SET LOCAL search_path = "%s"`, schema)

	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("set tenant schema: %w", err)
	}

	return nil
}

func validateSchemaName(schema string) error {

	if !schemaRegex.MatchString(schema) {
		return fmt.Errorf(
			"invalid schema name",
		)
	}

	return nil
}

/*
Transaction manager only for queryies.
Read only transaction - less blocking (?)
Used to set tenant.
*/
func (m *TxManager) WithinTransactionReadonly(
	ctx context.Context,
	tenantSchema string,
	fn func(q *querydb.Queries) error,
) error {

	tx, err := m.pool.BeginTx(
		ctx,
		pgx.TxOptions{
			AccessMode: pgx.ReadOnly,
		},
	)
	if err != nil {
		return fmt.Errorf("begin readonly tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := setTenantSchema(
		ctx,
		tx,
		tenantSchema,
	); err != nil {
		return err
	}

	q := querydb.New(tx)

	return fn(q)
}