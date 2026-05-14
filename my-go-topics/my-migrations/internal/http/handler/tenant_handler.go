package handler

import (
	"context"
	"errors"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
)

func (s *Server) CreateTenant(
	ctx context.Context,
	req httpapi.CreateTenantRequestObject,
) (httpapi.CreateTenantResponseObject, error) {

	id, err := s.createTenant.Handle(ctx, appcommand.CreateTenantCommand{
		SchemaName: req.Body.SchemaName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return httpapi.CreateTenant409JSONResponse{N409JSONResponse: conflictError()}, nil
		}
		return httpapi.CreateTenant500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CreateTenant201JSONResponse{Id: id}, nil
}

func (s *Server) GetTenant(
	ctx context.Context,
	req httpapi.GetTenantRequestObject,
) (httpapi.GetTenantResponseObject, error) {

	result, err := s.getTenant.Handle(ctx, appquery.GetTenantQuery{
		ID: req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
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
