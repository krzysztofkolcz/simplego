package events

import "github.com/krzysztofkolcz/my-go-rest-hex/internal/domain/events"

type EventHandler[T any] interface {
	Handle(event T) error
}

type EventBus struct {
	userCreated []EventHandler[events.UserCreated]
}

func NewEventBus() *EventBus {
	return &EventBus{}
}

func (b *EventBus) RegisterUserCreated(h EventHandler[events.UserCreated]) {
	b.userCreated = append(b.userCreated, h)
}

func (b *EventBus) PublishUserCreated(e events.UserCreated) error {
	for _, h := range b.userCreated {
		if err := h.Handle(e); err != nil {
			return err
		}
	}
	return nil
}
