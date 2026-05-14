package port

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
)

type CreateTodoPort interface {
	Handle(ctx context.Context, cmd appcommand.CreateTodoCommand) (uuid.UUID, error)
}

type CompleteTodoPort interface {
	Handle(ctx context.Context, cmd appcommand.CompleteTodoCommand) error
}

type DeleteTodoPort interface {
	Handle(ctx context.Context, cmd appcommand.DeleteTodoCommand) error
}

type GetTodoPort interface {
	Handle(ctx context.Context, q appquery.GetTodoQuery) (*appquery.GetTodoResult, error)
}

type ListTodosPort interface {
	Handle(ctx context.Context, q appquery.ListTodosQuery) ([]appquery.TodoResult, error)
}
