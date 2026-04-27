package integration

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func schemaExists(ctx context.Context, db *pgxpool.Pool, schema string) (bool, error) {
	var exists bool

	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.schemata
			WHERE schema_name = $1
		)
	`, schema).Scan(&exists)

	return exists, err
}

func tableExists(ctx context.Context, db *pgxpool.Pool, schema, table string) (bool, error) {
	var exists bool

	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = $1
			  AND table_name = $2
		)
	`, schema, table).Scan(&exists)

	return exists, err
}