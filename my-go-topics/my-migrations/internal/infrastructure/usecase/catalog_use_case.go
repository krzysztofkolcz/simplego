package usecase

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	"github.com/krzysztofkolcz/mymigrations/internal/application/port"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	catalogrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant/catalog"
)

type createComponentUseCase struct {
	txManager      *db.TxManager
	eventPublisher domain.EventPublisher
}

func NewCreateComponentUseCase(txManager *db.TxManager, publisher domain.EventPublisher) port.CreateComponentPort {
	return &createComponentUseCase{txManager: txManager, eventPublisher: publisher}
}

func (u *createComponentUseCase) Handle(ctx context.Context, cmd appcommand.CreateComponentCommand) (uuid.UUID, error) {
	uow := catalogrepo.NewCatalogUnitOfWork(u.txManager, cmd.TenantSchema)
	return appcommand.NewCreateComponentHandler(uow, u.eventPublisher).Handle(ctx, cmd)
}

type updateComponentUseCase struct {
	txManager      *db.TxManager
	eventPublisher domain.EventPublisher
}

func NewUpdateComponentUseCase(txManager *db.TxManager, publisher domain.EventPublisher) port.UpdateComponentPort {
	return &updateComponentUseCase{txManager: txManager, eventPublisher: publisher}
}

func (u *updateComponentUseCase) Handle(ctx context.Context, cmd appcommand.UpdateComponentCommand) error {
	uow := catalogrepo.NewCatalogUnitOfWork(u.txManager, cmd.TenantSchema)
	return appcommand.NewUpdateComponentHandler(uow, u.eventPublisher).Handle(ctx, cmd)
}

type getComponentUseCase struct {
	txManager *db.TxManager
}

func NewGetComponentUseCase(txManager *db.TxManager) port.GetComponentPort {
	return &getComponentUseCase{txManager: txManager}
}

func (u *getComponentUseCase) Handle(ctx context.Context, q appquery.GetComponentQuery) (*appquery.GetComponentResult, error) {
	repo := catalogrepo.NewComponentReadRepository(u.txManager, q.TenantSchema)
	return appquery.NewGetComponentHandler(repo).Handle(ctx, q)
}

type createProductUseCase struct {
	txManager      *db.TxManager
	eventPublisher domain.EventPublisher
}

func NewCreateProductUseCase(txManager *db.TxManager, publisher domain.EventPublisher) port.CreateProductPort {
	return &createProductUseCase{txManager: txManager, eventPublisher: publisher}
}

func (u *createProductUseCase) Handle(ctx context.Context, cmd appcommand.CreateProductCommand) (uuid.UUID, error) {
	uow := catalogrepo.NewCatalogUnitOfWork(u.txManager, cmd.TenantSchema)
	return appcommand.NewCreateProductHandler(uow, u.eventPublisher).Handle(ctx, cmd)
}

type getProductUseCase struct {
	txManager *db.TxManager
}

func NewGetProductUseCase(txManager *db.TxManager) port.GetProductPort {
	return &getProductUseCase{txManager: txManager}
}

func (u *getProductUseCase) Handle(ctx context.Context, q appquery.GetProductQuery) (*appquery.GetProductResult, error) {
	repo := catalogrepo.NewProductReadRepository(u.txManager, q.TenantSchema)
	return appquery.NewGetProductHandler(repo).Handle(ctx, q)
}

type setProductComponentUseCase struct {
	txManager      *db.TxManager
	eventPublisher domain.EventPublisher
}

func NewSetProductComponentUseCase(txManager *db.TxManager, eventPublisher domain.EventPublisher) port.SetProductComponentPort {

	return &setProductComponentUseCase{txManager: txManager, eventPublisher: eventPublisher}
}

func (u *setProductComponentUseCase) Handle(ctx context.Context, q appcommand.SetProductComponentCommand) error {
	uow := catalogrepo.NewCatalogUnitOfWork(u.txManager, q.TenantSchema)
	return appcommand.NewSetProductComponentHandler(uow, u.eventPublisher).Handle(ctx, q)
}


type removeProductComponentUseCase struct {
	txManager      *db.TxManager
	eventPublisher domain.EventPublisher
}

func NewRemoveProductComponentUseCase(txManager *db.TxManager, eventPublisher domain.EventPublisher) port.RemoveProductComponentPort {

	return &removeProductComponentUseCase{txManager: txManager, eventPublisher: eventPublisher}
}

func (u *removeProductComponentUseCase) Handle(ctx context.Context, q appcommand.RemoveProductComponentCommand) error {
	uow := catalogrepo.NewCatalogUnitOfWork(u.txManager, q.TenantSchema)
	return appcommand.NewRemoveProductComponentHandler(uow, u.eventPublisher).Handle(ctx, q)
}
