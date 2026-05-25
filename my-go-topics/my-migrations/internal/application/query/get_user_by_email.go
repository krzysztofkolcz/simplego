package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
)

type GetUserByEmailQuery struct {
	Email string
}

type GetUserByEmailResult struct {
	ID    uuid.UUID
	Email string
}

type GetUserByEmailHandler struct {
	repo user.Repository
}

func NewGetUserByEmailHandler(repo user.Repository) *GetUserByEmailHandler {
	return &GetUserByEmailHandler{repo: repo}
}

func (h *GetUserByEmailHandler) Handle(
	ctx context.Context,
	q GetUserByEmailQuery,
) (*GetUserByEmailResult, error) {

	email, err := user.NewEmail(q.Email)
	if err != nil {
		return nil, err
	}

	u, err := h.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &GetUserByEmailResult{
		ID:    u.ID,
		Email: u.Email.String(),
	}, nil
}
