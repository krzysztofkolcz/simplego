package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
)

type ListTodosQuery struct {}


type TodoResult struct {
	ID        uuid.UUID
	Title     string
	Completed bool
}

type ListTodosHandler struct{
	repo todo.ReadRepository
}

func NewListTodosHandler(repo todo.ReadRepository) *ListTodosHandler {
	return &ListTodosHandler{
		repo: repo,
	}
}

func (h *ListTodosHandler) Handle(ctx context.Context, query ListTodosQuery) ([]TodoResult, error) {

	list, err := h.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	response := []TodoResult{}
	for _,l := range list {
		res := TodoResult{
			ID:        l.ID,
			Title:     l.Title,
			Completed: l.Completed,
		}
		response = append(response, res)
	}
	return response, nil

}
