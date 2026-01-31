package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofkolcz/my-go-rest-foundation-002/internal/http/handlers"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)
	r.Get("/users", handlers.Users)
	r.Get("/users/{id}", handlers.GetUser)

	return r
}

