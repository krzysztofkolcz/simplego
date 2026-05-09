package public

import (
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

type TenantRepository struct {
	commandQ *commanddb.Queries
	queryQ   *querydb.Queries
}