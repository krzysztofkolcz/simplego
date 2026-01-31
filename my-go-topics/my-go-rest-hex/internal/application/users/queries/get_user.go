package users

import "github.com/krzysztofkolcz/my-go-rest-hex/internal/domain"

type GetUserQuery struct {
	ID string
}

type UserView struct {
	ID    string
	Email string
}

type GetUserHandler struct {
	Users domain.UserRepository
}

func (h *GetUserHandler) Handle(q GetUserQuery) (*UserView, error) {
	u, err := h.Users.GetByID(q.ID)
	if err != nil {
		return nil, err
	}

	return &UserView{
		ID:    u.ID,
		Email: u.Email,
	}, nil
}
