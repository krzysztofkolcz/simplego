package public

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
)

type UserRepository struct {
	commandQ *commanddb.Queries
	queryQ   *querydb.Queries
}

func NewUserRepository(
	commandQ *commanddb.Queries,
	queryQ *querydb.Queries,
) *UserRepository {

	return &UserRepository{
		commandQ: commandQ,
		queryQ:   queryQ,
	}
}

func (r *UserRepository) Create(
	ctx context.Context,
	u user.User,
) error {

	_, err := r.commandQ.CreateUser(
		ctx,
		commanddb.CreateUserParams{
			ID:    u.ID,
			Email: u.Email,
		},
	)

	return err
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*user.User, error) {

	row, err := r.queryQ.GetUserID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user.User{
		ID:    row.ID,
		Email: row.Email,
	}, nil
}
