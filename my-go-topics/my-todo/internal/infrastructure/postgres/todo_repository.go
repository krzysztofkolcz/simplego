// Package postgres is the outbound adapter that implements domain.TodoRepository.
//
// In Hexagonal Architecture terms this package is an ADAPTER: it sits at the
// boundary of the application and translates between the domain model and the
// database representation produced by sqlc.
//
// Key responsibilities:
//   - Call sqlc-generated query functions (never write raw SQL here)
//   - Map pgtype values to/from domain types
//   - Translate infrastructure errors into domain errors
//     (e.g., pgx.ErrNoRows → domain.ErrTodoNotFound)
//
// The domain package never imports this package. The dependency arrow points
// inward: postgres imports domain, not the other way around.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/C5383717/my-todo/internal/domain"
	"github.com/C5383717/my-todo/internal/infrastructure/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TodoRepository is the concrete implementation of domain.TodoRepository.
// It wraps the sqlc-generated Queries struct.
type TodoRepository struct {
	queries *sqlcdb.Queries
}

// NewTodoRepository creates a TodoRepository from a pgxpool connection pool.
// The pool is injected — this adapter never opens its own connection.
func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{
		queries: sqlcdb.New(pool),
	}
}

// Create persists a new Todo.
// Maps domain types → sqlc param types before calling the generated query.
func (r *TodoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	err := r.queries.CreateTodo(ctx, sqlcdb.CreateTodoParams{
		ID:          uuidToPgtype(todo.ID),
		Title:       todo.Title,
		Description: strPtrToPgText(todo.Description),
		Status:      sqlcdb.TodoStatus(todo.Status),
		CreatedAt:   timeToPgTimestamptz(todo.CreatedAt),
		UpdatedAt:   timeToPgTimestamptz(todo.UpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("todo repository: create: %w", err)
	}

	return nil
}

// GetByID retrieves a Todo by UUID.
// Translates pgx.ErrNoRows into domain.ErrTodoNotFound so callers can use
// errors.Is(err, domain.ErrTodoNotFound) without knowing about pgx.
func (r *TodoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Todo, error) {
	row, err := r.queries.GetTodoByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("todo repository: get by id: %w", domain.ErrTodoNotFound)
		}

		return nil, fmt.Errorf("todo repository: get by id: %w", err)
	}

	return rowToDomain(row), nil
}

// List returns all Todos ordered by creation time (newest first).
func (r *TodoRepository) List(ctx context.Context) ([]domain.Todo, error) {
	rows, err := r.queries.ListTodos(ctx)
	if err != nil {
		return nil, fmt.Errorf("todo repository: list: %w", err)
	}

	todos := make([]domain.Todo, 0, len(rows))
	for _, row := range rows {
		todos = append(todos, *rowToDomain(row))
	}

	return todos, nil
}

// Update persists the current state of a mutated aggregate.
// The caller (command handler) is responsible for mutating the aggregate
// before passing it here — this method just writes what it receives.
func (r *TodoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	err := r.queries.UpdateTodo(ctx, sqlcdb.UpdateTodoParams{
		ID:          uuidToPgtype(todo.ID),
		Title:       todo.Title,
		Description: strPtrToPgText(todo.Description),
		UpdatedAt:   timeToPgTimestamptz(todo.UpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("todo repository: update: %w", err)
	}

	return nil
}

// Delete removes a Todo permanently.
func (r *TodoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteTodo(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("todo repository: delete: %w", err)
	}

	return nil
}

// --- mapping helpers ---

// rowToDomain converts a sqlc-generated Todo row to a domain.Todo aggregate.
// Uses domain.Reconstitute — the only path to create a Todo from persisted data.
func rowToDomain(row sqlcdb.Todo) *domain.Todo {
	var desc *string
	if row.Description.Valid {
		s := row.Description.String
		desc = &s
	}

	status, _ := domain.EnsureValidStatus(string(row.Status))

	return domain.Reconstitute(
		pgUUIDToUUID(row.ID),
		row.Title,
		desc,
		status,
		row.CreatedAt.Time,
		row.UpdatedAt.Time,
	)
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgUUIDToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

func timeToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func strPtrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{String: *s, Valid: true}
}
