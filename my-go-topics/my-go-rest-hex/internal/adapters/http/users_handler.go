package http

import (
	"encoding/json"
	"net/http"

	"github.com/krzysztofkolcz/my-go-rest-hex/internal/application/users"
)

type UserHandlers struct {
	CreateUser *users.CreateUser
}

func (h *UserHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}

	user, err := h.CreateUser.Execute(users.CreateUserInput{
		Email: req.Email,
	})

	if err != nil {
		mapError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}
