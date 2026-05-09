package commanddb

// DB returns the underlying DBTX (pool, connection, or transaction).
// Used when the same connection must be shared across command and query instances,
// e.g. to create a querydb.Queries from the same transaction.
func (q *Queries) DB() DBTX {
	return q.db
}
