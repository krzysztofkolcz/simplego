package repository

import (
	"context"
	"fmt"

	"github.com/example/my-supertokens/internal/domain"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
)

// SuperTokensAuthService implements domain.AuthService using SuperTokens.
type SuperTokensAuthService struct{}

func NewSuperTokensAuthService() *SuperTokensAuthService {
	return &SuperTokensAuthService{}
}

func (s *SuperTokensAuthService) Register(_ context.Context, email, password string) (*domain.User, error) {
	result, err := emailpassword.SignUp("public", email, password)
	if err != nil {
		return nil, fmt.Errorf("sign up: %w", err)
	}
	if result.EmailAlreadyExistsError != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	return &domain.User{
		ID:    result.OK.User.ID,
		Email: result.OK.User.Email,
	}, nil
}

func (s *SuperTokensAuthService) Login(_ context.Context, email, password string) (*domain.User, error) {
	result, err := emailpassword.SignIn("public", email, password)
	if err != nil {
		return nil, fmt.Errorf("sign in: %w", err)
	}
	if result.WrongCredentialsError != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return &domain.User{
		ID:    result.OK.User.ID,
		Email: result.OK.User.Email,
	}, nil
}
