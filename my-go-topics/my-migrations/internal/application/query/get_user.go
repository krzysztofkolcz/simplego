package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
)

type GetUserQuery struct {
	ID uuid.UUID
}

type GetUserResult struct {
	ID    uuid.UUID
	Email string
}

type GetUserHandler struct {
	repo user.Repository
}

func NewGetUserHandler(repo user.Repository) *GetUserHandler {
	return &GetUserHandler{repo: repo}
}

func (h *GetUserHandler) Handle(
	ctx context.Context,
	q GetUserQuery,
) (*GetUserResult, error) {

	u, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &GetUserResult{
		ID:    u.ID,
		Email: u.Email,
	}, nil
}
