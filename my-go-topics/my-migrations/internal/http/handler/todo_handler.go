package handler

import (
	"context"
	"errors"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
	tenantrepo "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/repository/tenant"
)

func (s *Server) CreateTodo(
	ctx context.Context,
	req httpapi.CreateTodoRequestObject,
) (httpapi.CreateTodoResponseObject, error) {

	uow := tenantrepo.NewTodoUnitOfWork(s.txManager, tenantSchema(req.Params.XTenantID))
	id, err := appcommand.NewCreateTodoHandler(uow).Handle(ctx, appcommand.CreateTodoCommand{
		Title: req.Body.Title,
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

	repo := tenantrepo.NewTodoReadRepository(s.txManager, tenantSchema(req.Params.XTenantID))
	result, err := appquery.NewGetTodoHandler(repo).Handle(ctx, appquery.GetTodoQuery{
		ID: req.Id,
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

	uow := tenantrepo.NewTodoUnitOfWork(s.txManager, tenantSchema(req.Params.XTenantID))
	err := appcommand.NewCompleteTodoHandler(uow).Handle(ctx, appcommand.CompleteTodoCommand{
		ID: req.Id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httpapi.CompleteTodo404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.CompleteTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CompleteTodo204Response{}, nil
}

func (s *Server) DeleteTodo(
	ctx context.Context,
	req httpapi.DeleteTodoRequestObject,
) (httpapi.DeleteTodoResponseObject, error) {

	uow := tenantrepo.NewTodoUnitOfWork(s.txManager, tenantSchema(req.Params.XTenantID))
	err := appcommand.NewDeleteTodoHandler(uow).Handle(ctx, appcommand.DeleteTodoCommand{
		ID: req.Id,
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

	repo := tenantrepo.NewTodoReadRepository(s.txManager, tenantSchema(req.Params.XTenantID))
	list, err := appquery.NewListTodosHandler(repo).Handle(ctx, appquery.ListTodosQuery{})
	if err != nil {
		return httpapi.ListTodos500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	var responses = httpapi.ListTodos200JSONResponse{}
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