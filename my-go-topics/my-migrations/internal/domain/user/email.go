package user

import (
	"strings"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
)

type Email struct {
	value string
}

func NewEmail(s string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	parts := strings.SplitN(normalized, "@", 3)
	if len(parts) != 2 || parts[0] == "" || !strings.Contains(parts[1], ".") || strings.HasSuffix(parts[1], ".") {
		return Email{}, domain.ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

func (e Email) String() string {
	return e.value
}
