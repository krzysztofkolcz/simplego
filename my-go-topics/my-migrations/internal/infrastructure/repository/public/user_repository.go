package public

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command/public"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query/public"
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
			Email: u.Email.String(),
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}

	return nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*user.User, error) {

	row, err := r.queryQ.GetUserID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	email, err := user.NewEmail(row.Email)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:    row.ID,
		Email: email,
	}, nil
}

func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email user.Email,
) (*user.User, error) {

	row, err := r.queryQ.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	domainEmail, err := user.NewEmail(row.Email)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:    row.ID,
		Email: domainEmail,
	}, nil
}
