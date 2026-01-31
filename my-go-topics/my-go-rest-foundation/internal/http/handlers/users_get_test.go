package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/krzysztofkolcz/my-go-rest-foundation/internal/domain"
	"github.com/krzysztofkolcz/my-go-rest-foundation/internal/http/handlers"
)

func TestGetUser_OK(t *testing.T) {
	mockService := &MockUserService{
		GetByIDFn: func(id string) (*domain.User, error) {
			return &domain.User{
				ID:    "1",
				Email: "john@example.com",
			}, nil
		},
	}

	handler := handlers.NewUserHandlers(mockService)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req = req.WithContext(context.Background())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	mockService := &MockUserService{
		GetByIDFn: func(id string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	handler := handlers.NewUserHandlers(mockService)

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateUser_OK(t *testing.T) {
	mockService := &MockUserService{
		CreateFn: func(email string) (*domain.User, error) {
			return &domain.User{
				ID:    "3",
				Email: email,
			}, nil
		},
	}

	handler := handlers.NewUserHandlers(mockService)

	body := strings.NewReader(`{"email":"new@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	mockService := &MockUserService{
		CreateFn: func(email string) (*domain.User, error) {
			return nil, domain.ErrInvalidEmail
		},
	}

	handler := handlers.NewUserHandlers(mockService)

	body := strings.NewReader(`{"email":""}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)

	rec := httptest.NewRecorder()

	handler.CreateUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
