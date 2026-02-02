package events

import "time"

type UserCreated struct {
	UserID string
	Email  string
	At     time.Time
}

