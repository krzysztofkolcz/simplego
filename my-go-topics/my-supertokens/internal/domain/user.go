package domain

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// User represents an authenticated user.
type User struct {
	ID    string
	Email string
}
