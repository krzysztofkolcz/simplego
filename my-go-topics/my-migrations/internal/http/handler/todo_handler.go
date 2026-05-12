package handler

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
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
		return httpapi.CreateTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.CreateTodo201JSONResponse{Id: id}, nil
}

func (s *Server) GetTodo(
	ctx context.Context,
	req httpapi.GetTodoRequestObject,
) (httpapi.GetTodoResponseObject, error) {

	var result *appquery.GetTodoResult

	err := s.txManager.WithinTransactionReadonly(
		ctx,
		tenantSchema(req.Params.XTenantID),
		func(q *querydb.Queries) error {
			repo := tenantrepo.NewTodoRepository(nil, q)
			var err error
			result, err = appquery.NewGetTodoHandler(repo).Handle(ctx, appquery.GetTodoQuery{
				ID: req.Id,
			})
			return err
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.GetTodo404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.GetTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.GetTodo200JSONResponse{
		Id:        result.ID,
		Title:     result.Title,
		Completed: result.Completed,
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
		if errors.Is(err, pgx.ErrNoRows) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.DeleteTodo404JSONResponse{N404JSONResponse: notFoundError()}, nil
		}
		return httpapi.DeleteTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	return httpapi.DeleteTodo204Response{}, nil
}


func (s *Server) ListTodos(
	ctx context.Context,
	req httpapi.ListTodosRequestObject,
) (httpapi.ListTodosResponseObject, error){

	var list = []appquery.TodoResult{}

	err := s.txManager.WithinTransactionReadonly(
		ctx,
		tenantSchema(req.Params.XTenantID),
		func(q *querydb.Queries) error {
			var err error
			repo := tenantrepo.NewTodoRepository(nil, q)
			list, err = appquery.NewListTodosHandler(repo).Handle(ctx, appquery.ListTodosQuery{})
			return err
		},
	)

	
	if err != nil {
		return httpapi.ListTodos500JSONResponse{N500JSONResponse: internalError(err)}, nil
	}

	var responses = httpapi.ListTodos200JSONResponse{}
	for _,l:= range list{
		res := httpapi.Todo{
			Completed: l.Completed,
			CreatedAt: nil,
			Id:        l.ID,
			Title:     l.Title,
		}
		responses = append(responses, res)
	}
	return responses, nil
}