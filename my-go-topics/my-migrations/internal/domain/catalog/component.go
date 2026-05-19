package catalog

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
)

type Component struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Code            string
	Manufacturer    string
	ManufacturerURL string
	Price           float64
	WeightG         int
	Quantity        int
	ImageURL        string
	CreatedAt       time.Time

	events []domain.DomainEvent
}

func NewComponent(id uuid.UUID, name, code string) (*Component, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidName
	}
	if strings.TrimSpace(code) == "" {
		return nil, ErrInvalidCode
	}
	return &Component{
		ID:        id,
		Name:      strings.TrimSpace(name),
		Code:      strings.TrimSpace(code),
		CreatedAt: time.Now(),
	}, nil
}

func (c *Component) PullEvents() []domain.DomainEvent {
	events := c.events
	c.events = nil
	return events
}
