package users

import "github.com/krzysztofkolcz/my-go-rest-hex/internal/domain"

// type CreateUser struct {
// 	Users domain.UserRepository
// }

// CreateUserCommand
type CreateUserInput struct {
	Email string
}

type CreateUserOutput struct {
	ID    string
	Email string
}


// domain  (domain.User) nie zwaracamy
// func (uc *CreateUser) Execute(input CreateUserInput) (CreateUserOutput, error) {
// 	if input.Email == "" {
// 		return CreateUserOutput{}, domain.ErrInvalidEmail
// 	}

// 	if _, err := uc.Users.GetByEmail(input.Email); err == nil {
// 		return CreateUserOutput{}, domain.ErrUserExists
// 	}

// 	user, err := domain.NewUser(generateID(), input.Email)
// 	if err != nil {
// 		return CreateUserOutput{}, err
// 	}

// 	if err := uc.Users.Save(user); err != nil {
// 		return CreateUserOutput{}, err
// 	}

// 	return CreateUserOutput{
// 		ID:    user.ID,
// 		Email: user.Email,
// 	}, nil
// }

type CreateUser struct {
	Users domain.UserRepository
	Tx    TransactionManager
}

func (uc *CreateUser) Execute(input CreateUserInput) (CreateUserOutput, error) {
	var result CreateUserOutput

	err := uc.Tx.WithinTransaction(func() error {
		user, err := domain.NewUser(generateID(), input.Email)
		if err != nil {
			return err
		}

		if err := uc.Users.Save(user); err != nil {
			return err
		}

		result = CreateUserOutput{
			ID:    user.ID,
			Email: user.Email,
		}
		return nil
	})

	return result, err
}
