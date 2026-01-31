package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	userscommands "github.com/krzysztofkolcz/my-go-rest-hex/internal/application/users/commands"
	usersqueries "github.com/krzysztofkolcz/my-go-rest-hex/internal/application/users/queries"
)

type UserHandlers struct {
	CreateUser *userscommands.CreateUserHandler
	GetUser *usersqueries.GetUserHandler
}

func (h *UserHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	err := h.CreateUser.Handle(userscommands.CreateUserCommand{
		Email: req.Email,
	})

	if err != nil {
		mapError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandlers) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	view, err := h.GetUser.Handle(usersqueries.GetUserQuery{
		ID: id,
	})

	if err != nil {
		mapError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, view)
}