package usecase

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/application/port"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	tenantrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant"
)

type createTodoUseCase struct {
	txManager      *db.TxManager
	eventPublisher domain.EventPublisher
}

func NewCreateTodoUseCase(txManager *db.TxManager, publisher domain.EventPublisher) port.CreateTodoPort {
	return &createTodoUseCase{txManager: txManager, eventPublisher: publisher}
}

func (u *createTodoUseCase) Handle(ctx context.Context, cmd appcommand.CreateTodoCommand) (uuid.UUID, error) {
	uow := tenantrepo.NewTodoUnitOfWork(u.txManager, cmd.TenantSchema)
	return appcommand.NewCreateTodoHandler(uow, u.eventPublisher).Handle(ctx, cmd)
}

type completeTodoUseCase struct {
	txManager *db.TxManager
}

func NewCompleteTodoUseCase(txManager *db.TxManager) port.CompleteTodoPort {
	return &completeTodoUseCase{txManager: txManager}
}

func (u *completeTodoUseCase) Handle(ctx context.Context, cmd appcommand.CompleteTodoCommand) error {
	uow := tenantrepo.NewTodoUnitOfWork(u.txManager, cmd.TenantSchema)
	return appcommand.NewCompleteTodoHandler(uow).Handle(ctx, cmd)
}

type deleteTodoUseCase struct {
	txManager *db.TxManager
}

func NewDeleteTodoUseCase(txManager *db.TxManager) port.DeleteTodoPort {
	return &deleteTodoUseCase{txManager: txManager}
}

func (u *deleteTodoUseCase) Handle(ctx context.Context, cmd appcommand.DeleteTodoCommand) error {
	uow := tenantrepo.NewTodoUnitOfWork(u.txManager, cmd.TenantSchema)
	return appcommand.NewDeleteTodoHandler(uow).Handle(ctx, cmd)
}

type getTodoUseCase struct {
	txManager *db.TxManager
}

func NewGetTodoUseCase(txManager *db.TxManager) port.GetTodoPort {
	return &getTodoUseCase{txManager: txManager}
}

func (u *getTodoUseCase) Handle(ctx context.Context, q appquery.GetTodoQuery) (*appquery.GetTodoResult, error) {
	repo := tenantrepo.NewTodoReadRepository(u.txManager, q.TenantSchema)
	return appquery.NewGetTodoHandler(repo).Handle(ctx, q)
}

type listTodosUseCase struct {
	txManager *db.TxManager
}

func NewListTodosUseCase(txManager *db.TxManager) port.ListTodosPort {
	return &listTodosUseCase{txManager: txManager}
}

func (u *listTodosUseCase) Handle(ctx context.Context, q appquery.ListTodosQuery) ([]appquery.TodoResult, error) {
	repo := tenantrepo.NewTodoReadRepository(u.txManager, q.TenantSchema)
	return appquery.NewListTodosHandler(repo).Handle(ctx, q)
}
