package user

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
)

func TestNewEmail_Valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"User@Example.COM",
		"  user@example.com  ",
		"user+tag@sub.domain.org",
	}
	for _, raw := range cases {
		email, err := NewEmail(raw)
		require.NoError(t, err, "input: %q", raw)
		require.NotEmpty(t, email.String())
	}
}

func TestNewEmail_Normalizes(t *testing.T) {
	email, err := NewEmail("  User@Example.COM  ")
	require.NoError(t, err)
	require.Equal(t, "user@example.com", email.String())
}

func TestNewEmail_Invalid(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"notanemail",
		"missing-at.com",
		"@nodomain",
		"no@dot",
		"trailing@dot.",
	}
	for _, raw := range cases {
		_, err := NewEmail(raw)
		require.ErrorIs(t, err, domain.ErrInvalidEmail, "input: %q", raw)
	}
}
