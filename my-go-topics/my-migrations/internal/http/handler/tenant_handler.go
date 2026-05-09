package handler

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	publicrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/public"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
)

func (s *Server) CreateTenant(
	ctx context.Context,
	req httpapi.CreateTenantRequestObject,
) (httpapi.CreateTenantResponseObject, error) {

	uow := publicrepo.NewTenantUnitOfWork(s.txManager)
	id, err := appcommand.NewCreateTenantHandler(uow).Handle(ctx, appcommand.CreateTenantCommand{
		SchemaName: req.Body.SchemaName,
	})
	if err != nil {
		return httpapi.CreateTenant500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CreateTenant201JSONResponse{Id: id}, nil
}

func (s *Server) GetTenant(
	ctx context.Context,
	req httpapi.GetTenantRequestObject,
) (httpapi.GetTenantResponseObject, error) {

	repo := publicrepo.NewTenantRepository(s.commandQ, s.queryQ)
	result, err := appquery.NewGetTenantHandler(repo).Handle(ctx, appquery.GetTenantQuery{
		ID: req.Id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.GetTenant404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.GetTenant500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.GetTenant200JSONResponse{
		Id:               result.ID,
		SchemaName:       result.SchemaName,
		MigrationStatus:  &result.MigrationStatus,
		MigrationError:   &result.MigrationError,
		MigrationUpdated: &result.MigrationUpdated,
		CreatedAt:        &result.CreatedAt,
	}, nil
}
