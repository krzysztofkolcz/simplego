package domain

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(id string) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) List() ([]User, error) {
	return s.repo.List()
}

func (s *UserService) Create(email string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}

	if _, err := s.repo.GetByEmail(email); err == nil {
		return nil, ErrUserExists
	}

	user := User{
		ID:    generateID(), // helper
		Email: email,
	}

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	return &user, nil
}