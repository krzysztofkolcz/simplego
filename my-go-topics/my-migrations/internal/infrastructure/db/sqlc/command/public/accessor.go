package commanddb

func (q *Queries) DB() DBTX {
	return q.db
}
