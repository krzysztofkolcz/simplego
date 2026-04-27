package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krzysztofkolcz/my-sqlc/internal/db"
)

func WithTenant(ctx context.Context, pool *pgxpool.Pool, schema string) (*db.Queries, func(), error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}

	_, err = conn.Exec(ctx,
		fmt.Sprintf(`SET search_path TO %s, public`, schema),
	)
	if err != nil {
		conn.Release()
		return nil, nil, err
	}

	q := db.New(conn)

	cleanup := func() {
		conn.Release()
	}

	return q, cleanup, nil
}