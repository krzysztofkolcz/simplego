package handler

import (
	"context"
	"errors"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
)

func (s *Server) CreateTodo(
	ctx context.Context,
	req httpapi.CreateTodoRequestObject,
) (httpapi.CreateTodoResponseObject, error) {

	id, err := s.createTodo.Handle(ctx, appcommand.CreateTodoCommand{
		TenantSchema: tenantSchema(req.Params.XTenantID),
		Title:        req.Body.Title,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTitle) {
			return httpapi.CreateTodo400JSONResponse{N400JSONResponse: badRequestError(err.Error())}, nil
		}
		return httpapi.CreateTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CreateTodo201JSONResponse{Id: id}, nil
}

func (s *Server) GetTodo(
	ctx context.Context,
	req httpapi.GetTodoRequestObject,
) (httpapi.GetTodoResponseObject, error) {

	result, err := s.getTodo.Handle(ctx, appquery.GetTodoQuery{
		TenantSchema: tenantSchema(req.Params.XTenantID),
		ID:           req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.GetTodo404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.GetTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.GetTodo200JSONResponse{
		Id:        result.ID,
		Title:     result.Title,
		Completed: result.Completed,
		CreatedAt: &result.CreatedAt,
	}, nil
}

func (s *Server) CompleteTodo(
	ctx context.Context,
	req httpapi.CompleteTodoRequestObject,
) (httpapi.CompleteTodoResponseObject, error) {

	err := s.completeTodo.Handle(ctx, appcommand.CompleteTodoCommand{
		TenantSchema: tenantSchema(req.Params.XTenantID),
		ID:           req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.CompleteTodo404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		if errors.Is(err, domain.ErrAlreadyCompleted) {
			return httpapi.CompleteTodo409JSONResponse{N409JSONResponse: httpapi.N409JSONResponse{
				Error: httpapi.DetailedError{Status: 409, Code: "CONFLICT", Message: err.Error()},
			}}, nil
		}
		return httpapi.CompleteTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CompleteTodo204Response{}, nil
}

func (s *Server) DeleteTodo(
	ctx context.Context,
	req httpapi.DeleteTodoRequestObject,
) (httpapi.DeleteTodoResponseObject, error) {

	err := s.deleteTodo.Handle(ctx, appcommand.DeleteTodoCommand{
		TenantSchema: tenantSchema(req.Params.XTenantID),
		ID:           req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.DeleteTodo404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.DeleteTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.DeleteTodo204Response{}, nil
}

func (s *Server) ListTodos(
	ctx context.Context,
	req httpapi.ListTodosRequestObject,
) (httpapi.ListTodosResponseObject, error) {

	list, err := s.listTodos.Handle(ctx, appquery.ListTodosQuery{
		TenantSchema: tenantSchema(req.Params.XTenantID),
	})
	if err != nil {
		return httpapi.ListTodos500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	responses := httpapi.ListTodos200JSONResponse{}
	for _, l := range list {
		responses = append(responses, httpapi.Todo{
			Completed: l.Completed,
			CreatedAt: &l.CreatedAt,
			Id:        l.ID,
			Title:     l.Title,
		})
	}
	return responses, nil
}
