package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofkolcz/my-go-rest-foundation-002/internal/domain"
	slogctx "github.com/veqryn/slog-context"
)

var mockUsers = []domain.User{
	{ID: "1", Email: "john@example.com"},
	{ID: "2", Email: "kate@example.com"},
}

func Users(w http.ResponseWriter, r *http.Request) {
	logger := slogctx.FromCtx(r.Context())
	logger.Info("list users")

	users := mockUsers

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}


func GetUser(w http.ResponseWriter, r *http.Request) {
	logger := slogctx.FromCtx(r.Context())

	id := chi.URLParam(r, "id")
	logger = logger.With("user_id", id)

	for _, u := range mockUsers {
		if u.ID == id {
			logger.Info("user found")

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(u)
			return
		}
	}

	logger.Warn("user not found")
	http.Error(w, "user not found", http.StatusNotFound)
}