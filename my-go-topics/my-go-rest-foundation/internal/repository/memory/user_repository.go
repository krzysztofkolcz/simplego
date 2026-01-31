package memory

import (
	"sync"

	"github.com/krzysztofkolcz/my-go-rest-foundation/internal/domain"
)

type UserRepository struct {
	mu    sync.Mutex
	users []domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: []domain.User{
			{ID: "1", Email: "john@example.com"},
			{ID: "2", Email: "kate@example.com"},
		},
	}
}

func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *UserRepository) List() ([]domain.User, error) {
	return r.users, nil
}

func (r *UserRepository) Save(user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users = append(r.users, user)
	return nil
}
