package catalog

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       float64
	Tags        []string
	Recipe      []RecipeItem
	CreatedAt   time.Time

	events []domain.DomainEvent
}

func NewProduct(id uuid.UUID, name string) (*Product, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidName
	}
	return &Product{
		ID:        id,
		Name:      strings.TrimSpace(name),
		Tags:      []string{},
		Recipe:    []RecipeItem{},
		CreatedAt: time.Now(),
	}, nil
}

// AddComponent ustawia ilość komponentu w recepturze; nadpisuje jeśli już istnieje.
func (p *Product) AddComponent(componentID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	for i, item := range p.Recipe {
		if item.ComponentID == componentID {
			p.Recipe[i].Quantity = quantity
			return nil
		}
	}
	p.Recipe = append(p.Recipe, RecipeItem{ComponentID: componentID, Quantity: quantity})
	return nil
}

func (p *Product) RemoveComponent(componentID uuid.UUID) error {
	for i, item := range p.Recipe {
		if item.ComponentID == componentID {
			p.Recipe = append(p.Recipe[:i], p.Recipe[i+1:]...)
			return nil
		}
	}
	return ErrComponentNotFound
}

func (p *Product) PullEvents() []domain.DomainEvent {
	events := p.events
	p.events = nil
	return events
}
