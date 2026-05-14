package handler

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
)

func (s *Server) CreateUser(
	ctx context.Context,
	req httpapi.CreateUserRequestObject,
) (httpapi.CreateUserResponseObject, error) {

	id, err := s.createUser.Handle(ctx, appcommand.CreateUserCommand{
		Email: string(req.Body.Email),
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidEmail) {
			return httpapi.CreateUser400JSONResponse{N400JSONResponse: badRequestError(err.Error())}, nil
		}
		if errors.Is(err, domain.ErrConflict) {
			return httpapi.CreateUser409JSONResponse{N409JSONResponse: conflictError()}, nil
		}
		return httpapi.CreateUser500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CreateUser201JSONResponse{Id: id}, nil
}

func (s *Server) GetUser(
	ctx context.Context,
	req httpapi.GetUserRequestObject,
) (httpapi.GetUserResponseObject, error) {

	result, err := s.getUser.Handle(ctx, appquery.GetUserQuery{
		ID: req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.GetUser404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.GetUser500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.GetUser200JSONResponse{
		Id:    result.ID,
		Email: openapi_types.Email(result.Email),
	}, nil
}
