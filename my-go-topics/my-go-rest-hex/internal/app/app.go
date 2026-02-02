package app

import (
	users "github.com/krzysztofkolcz/my-go-rest-hex/internal/application/users/commands"
	"github.com/krzysztofkolcz/my-go-rest-hex/internal/repository/memory"
)

type App struct {
	CreateUser users.CreateUserHandler
}

func New() *App {
	userRepo := memory.NewUserRepository()
	userService := domain.NewUserService(userRepo)

	return &App{
		UserService: userService,
	}
}
