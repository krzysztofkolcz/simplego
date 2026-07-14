package domain

import "context"

// AuthService defines the authentication operations.
type AuthService interface {
	Register(ctx context.Context, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
}
