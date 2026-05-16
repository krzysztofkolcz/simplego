package db

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commanddbpub "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/public"
	commanddbtenant "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
	querydbpub "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/public"
	querydbtenant "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/tenant"
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
	fn func(q *commanddbtenant.Queries) error,
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

	q := commanddbtenant.New(tx)

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

func (m *TxManager) WithinPublicTransaction(
	ctx context.Context,
	fn func(q *commanddbpub.Queries) error,
) error {

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin public tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := commanddbpub.New(tx)

	if err := fn(q); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit public tx: %w", err)
	}

	return nil
}

func (m *TxManager) WithinPublicTransactionReadonly(
	ctx context.Context,
	fn func(q *querydbpub.Queries) error,
) error {

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("begin public readonly tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := querydbpub.New(tx)

	return fn(q)
}

/*
Transaction manager only for queryies.
Read only transaction - less blocking (?)
Used to set tenant.
*/
func (m *TxManager) WithinTransactionReadonly(
	ctx context.Context,
	tenantSchema string,
	fn func(q *querydbtenant.Queries) error,
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

	q := querydbtenant.New(tx)

	return fn(q)
}