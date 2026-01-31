package domain

type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	List() ([]User, error)
	Save(user User) error
}
