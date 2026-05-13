package user

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID
	Email Email
}

func NewUser(id uuid.UUID, email Email) *User {
	return &User{ID: id, Email: email}
}
