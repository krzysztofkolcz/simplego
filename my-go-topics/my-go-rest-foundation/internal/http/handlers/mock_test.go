package handlers_test

import "github.com/krzysztofkolcz/my-go-rest-foundation/internal/domain"

type MockUserService struct {
	GetByIDFn func(id string) (*domain.User, error)
	CreateFn  func(email string) (*domain.User, error)
}

func (m *MockUserService) List() []domain.User {
	return nil
}

func (m *MockUserService) GetByID(id string) (*domain.User, error) {
	return m.GetByIDFn(id)
}

func (m *MockUserService) Create(email string) (*domain.User, error) {
	return m.CreateFn(email)
}
