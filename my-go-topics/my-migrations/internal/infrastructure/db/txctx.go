package db

import (
	"context"

	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/tenant"
)

type txQueriesKey struct{}

func ContextWithTxQueries(ctx context.Context, q *commanddb.Queries) context.Context {
	return context.WithValue(ctx, txQueriesKey{}, q)
}

func TxQueriesFromCtx(ctx context.Context) (*commanddb.Queries, bool) {
	q, ok := ctx.Value(txQueriesKey{}).(*commanddb.Queries)
	return q, ok
}
