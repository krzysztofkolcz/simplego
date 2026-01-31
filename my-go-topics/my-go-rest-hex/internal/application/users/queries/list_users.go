package users

import "github.com/krzysztofkolcz/my-go-rest-hex/internal/domain"

type ListUsersHandler struct {
	Users domain.UserRepository
}

func (h *ListUsersHandler) Handle() ([]UserView, error) {
	users, err := h.Users.List()
	if err != nil {
		return nil, err
	}

	views := make([]UserView, 0, len(users))
	for _, u := range users {
		views = append(views, UserView{
			ID:    u.ID,
			Email: u.Email,
		})
	}

	return views, nil
}