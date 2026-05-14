package usecase

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/application/port"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/tenant"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
	publicrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/public"
)

type createTenantUseCase struct {
	txManager *db.TxManager
}

func NewCreateTenantUseCase(txManager *db.TxManager) port.CreateTenantPort {
	return &createTenantUseCase{txManager: txManager}
}

func (u *createTenantUseCase) Handle(ctx context.Context, cmd appcommand.CreateTenantCommand) (uuid.UUID, error) {
	uow := publicrepo.NewTenantUnitOfWork(u.txManager)
	return appcommand.NewCreateTenantHandler(uow).Handle(ctx, cmd)
}

type getTenantUseCase struct {
	repo tenant.Repository
}

func NewGetTenantUseCase(commandQ *commanddb.Queries, queryQ *querydb.Queries) port.GetTenantPort {
	return &getTenantUseCase{repo: publicrepo.NewTenantRepository(commandQ, queryQ)}
}

func (u *getTenantUseCase) Handle(ctx context.Context, q appquery.GetTenantQuery) (*appquery.GetTenantResult, error) {
	return appquery.NewGetTenantHandler(u.repo).Handle(ctx, q)
}
