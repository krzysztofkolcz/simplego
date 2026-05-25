package usecase

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/application/port"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/public"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/public"
	publicrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/public"
)

type createUserUseCase struct {
	txManager *db.TxManager
}

func NewCreateUserUseCase(txManager *db.TxManager) port.CreateUserPort {
	return &createUserUseCase{txManager: txManager}
}

func (u *createUserUseCase) Handle(ctx context.Context, cmd appcommand.CreateUserCommand) (uuid.UUID, error) {
	uow := publicrepo.NewUserUnitOfWork(u.txManager)
	return appcommand.NewCreateUserHandler(uow).Handle(ctx, cmd)
}

type getUserUseCase struct {
	repo user.Repository
}

func NewGetUserUseCase(commandQ *commanddb.Queries, queryQ *querydb.Queries) port.GetUserPort {
	return &getUserUseCase{repo: publicrepo.NewUserRepository(commandQ, queryQ)}
}

func (u *getUserUseCase) Handle(ctx context.Context, q appquery.GetUserQuery) (*appquery.GetUserResult, error) {
	return appquery.NewGetUserHandler(u.repo).Handle(ctx, q)
}

type getUserByEmailUseCase struct {
	repo user.Repository
}

func NewGetUserByEmailUseCase(commandQ *commanddb.Queries, queryQ *querydb.Queries) port.GetUserByEmailPort {
	return &getUserByEmailUseCase{repo: publicrepo.NewUserRepository(commandQ, queryQ)}
}

func (u *getUserByEmailUseCase) Handle(ctx context.Context, q appquery.GetUserByEmailQuery) (*appquery.GetUserByEmailResult, error) {
	return appquery.NewGetUserByEmailHandler(u.repo).Handle(ctx, q)
}
