package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/supertokens"
)

// NewRouter builds the Chi router with SuperTokens middleware wired in.
func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// SuperTokens middleware: handles /auth/* routes (login, signup, session refresh, etc.)
	// and injects CORS headers needed by the frontend SDK.
	r.Use(func(next http.Handler) http.Handler {
		return supertokens.Middleware(next)
	})

	// Public auth routes (custom — in addition to SuperTokens built-in /auth/* routes)
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/logout", session.VerifySession(nil, h.Logout))

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return session.VerifySession(nil, next.ServeHTTP)
		})
		r.Get("/me", h.Me)
	})

	return r
}
