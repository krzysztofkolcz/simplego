package todo

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
)

type Todo struct {
	ID        uuid.UUID
	Title     string
	Completed bool
	CreatedAt time.Time
}

func NewTodo(id uuid.UUID, title string) (*Todo, error) {
	if strings.TrimSpace(title) == "" {
		return nil, domain.ErrInvalidTitle
	}
	return &Todo{
		ID:    id,
		Title: strings.TrimSpace(title),
	}, nil
}

func (t *Todo) Complete() error {
	if t.Completed {
		return domain.ErrAlreadyCompleted
	}
	t.Completed = true
	return nil
}
