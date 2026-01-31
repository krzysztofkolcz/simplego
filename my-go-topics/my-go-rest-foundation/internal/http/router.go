package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	app "github.com/krzysztofkolcz/my-go-rest-foundation/internal"
	"github.com/krzysztofkolcz/my-go-rest-foundation/internal/http/handlers"
)

func NewRouter(app *app.App) http.Handler {
	r := chi.NewRouter()

	userHandlers := handlers.NewUserHandlers(app.UserService)

	r.Get("/health", handlers.Health)
	r.Get("/users/{id}", userHandlers.GetUser)
	r.Post("/users", userHandlers.CreateUser)

	return r
}
