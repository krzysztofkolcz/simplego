package domain

type User struct {
	ID    string
	Email string
}

func NewUser(id, email string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}
	return &User{ID: id, Email: email}, nil
}
