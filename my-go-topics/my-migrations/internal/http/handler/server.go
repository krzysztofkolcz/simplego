package handler

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

// Server implements httpapi.StrictServerInterface.
// Methods are split across tenant_handler.go, user_handler.go, todo_handler.go.
type Server struct {
	txManager      *db.TxManager
	queryQ         *querydb.Queries   // pool-backed, for public schema reads
	commandQ       *commanddb.Queries // pool-backed, for public schema reads
	eventPublisher domain.EventPublisher
}

func NewServer(txManager *db.TxManager, pool *pgxpool.Pool, eventPublisher domain.EventPublisher) *Server {
	return &Server{
		txManager:      txManager,
		queryQ:         querydb.New(pool),
		commandQ:       commanddb.New(pool),
		eventPublisher: eventPublisher,
	}
}
