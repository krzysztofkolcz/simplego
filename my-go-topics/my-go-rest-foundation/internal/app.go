package app

import (
	"github.com/krzysztofkolcz/my-go-rest-foundation/internal/domain"
	"github.com/krzysztofkolcz/my-go-rest-foundation/internal/repository/memory"
)

type App struct {
	UserService *domain.UserService
}

func New() *App {
	userRepo := memory.NewUserRepository()
	userService := domain.NewUserService(userRepo)

	return &App{
		UserService: userService,
	}
}
