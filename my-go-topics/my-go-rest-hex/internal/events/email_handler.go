package events

import (
	"fmt"

	"github.com/krzysztofkolcz/my-go-rest-hex/internal/domain/events"
)

type SendWelcomeEmail struct {
	// Mailer Mailer
}

func (h *SendWelcomeEmail) Handle(e events.UserCreated) error {
	// return h.Mailer.Send(
	// 	e.Email,
	// 	"Welcome!",
	// 	"Thanks for joining!",
	// )
	fmt.Print(e.Email)
	return nil
}