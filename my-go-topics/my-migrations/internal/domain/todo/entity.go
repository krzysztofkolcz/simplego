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

	events []domain.DomainEvent
}

func NewTodo(id uuid.UUID, title string) (*Todo, error) {
	if strings.TrimSpace(title) == "" {
		return nil, domain.ErrInvalidTitle
	}
	t := &Todo{
		ID:    id,
		Title: strings.TrimSpace(title),
	}
	t.record(TodoCreated{TodoID: id, Title: t.Title, OccurredAt: time.Now()})
	return t, nil
}

func (t *Todo) Complete() error {
	if t.Completed {
		return domain.ErrAlreadyCompleted
	}
	t.Completed = true
	return nil
}

// PullEvents returns recorded events and clears the internal list.
// Call after a successful transaction to publish events downstream.
func (t *Todo) PullEvents() []domain.DomainEvent {
	events := t.events
	t.events = nil
	return events
}

func (t *Todo) record(e domain.DomainEvent) {
	t.events = append(t.events, e)
}
