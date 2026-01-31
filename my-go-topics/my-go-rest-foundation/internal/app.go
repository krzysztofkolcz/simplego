package app

import "github.com/krzysztofkolcz/my-go-rest-foundation/internal/domain"

type App struct {
	UserService *domain.UserService
}

func New() *App {
	return &App{
		UserService: domain.NewUserService(),
	}
}
