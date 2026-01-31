package domain

type UserService interface {
	List() []User
	GetByID(id string) (*User, error)
	Create(email string) (*User, error)
}

type InMemoryUserService struct {
	users []User
}

func NewInMemoryUserService() *InMemoryUserService {
	return &InMemoryUserService{
		users: []User{
			{ID: "1", Email: "john@example.com"},
			{ID: "2", Email: "kate@example.com"},
		},
	}
}


func (s *InMemoryUserService) List() []User {
	return s.users
}

func (s *InMemoryUserService) GetByID(id string) (*User, error) {
	for _, u := range s.users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *InMemoryUserService) Create(email string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}

	for _, u := range s.users {
		if u.Email == email {
			return nil, ErrUserExists
		}
	}

	user := User{
		ID:    string(rune(len(s.users) + 1)),
		Email: email,
	}

	s.users = append(s.users, user)
	return &user, nil
}
