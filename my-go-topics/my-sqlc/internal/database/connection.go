package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(conn string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, err
	}
	return &DB{Pool: pool}, nil
}