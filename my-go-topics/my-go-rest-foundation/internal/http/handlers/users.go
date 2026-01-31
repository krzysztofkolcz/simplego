package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofkolcz/my-go-rest-foundation/internal/domain"
	apphttp "github.com/krzysztofkolcz/my-go-rest-foundation/internal/http/response"
)

type UserHandlers struct {
	Users *domain.UserService
}

func NewUserHandlers(users *domain.UserService) *UserHandlers {
	return &UserHandlers{Users: users}
}

type CreateUserRequest struct {
	Email string `json:"email"`
}

type CreateUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *UserHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.Users.GetByID(id)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			apphttp.JSON(w, http.StatusNotFound, apphttp.ErrorResponse{
				Error:   "user_not_found",
				Message: "user does not exist",
			})
		default:
			apphttp.JSON(w, http.StatusInternalServerError, apphttp.ErrorResponse{
				Error:   "internal_error",
				Message: "unexpected error",
			})
		}
		return
	}

	apphttp.JSON(w, http.StatusOK, user)
}

func (h *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		apphttp.JSON(w, http.StatusBadRequest, apphttp.ErrorResponse{
			Error:   "invalid_json",
			Message: "invalid request body",
		})
		return
	}

	user, err := h.Users.Create(req.Email)
	if err != nil {
		switch err {
		case domain.ErrInvalidEmail:
			apphttp.JSON(w, http.StatusBadRequest, apphttp.ErrorResponse{
				Error:   "invalid_email",
				Message: "email is invalid",
			})
		case domain.ErrUserExists:
			apphttp.JSON(w, http.StatusConflict, apphttp.ErrorResponse{
				Error:   "user_exists",
				Message: "user already exists",
			})
		default:
			apphttp.JSON(w, http.StatusInternalServerError, apphttp.ErrorResponse{
				Error:   "internal_error",
				Message: "unexpected error",
			})
		}
		return
	}

	apphttp.JSON(w, http.StatusCreated, user)
}
