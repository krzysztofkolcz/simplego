package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain/tenant"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
)

// ── Todo ──────────────────────────────────────────────────────────────────────

type fakeTodoUoW struct {
	repo todo.Repository
	err  error // returned by Execute before calling fn
}

func (f *fakeTodoUoW) Execute(_ context.Context, fn func(todo.Repository) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(f.repo)
}

type fakeTodoRepo struct {
	created   []todo.Todo
	completed []uuid.UUID
	deleted   []uuid.UUID
	err       error // returned by every method
}

func (f *fakeTodoRepo) Create(_ context.Context, t todo.Todo) error {
	if f.err != nil {
		return f.err
	}
	f.created = append(f.created, t)
	return nil
}

func (f *fakeTodoRepo) GetByID(_ context.Context, _ uuid.UUID) (*todo.Todo, error) {
	return nil, f.err
}

func (f *fakeTodoRepo) Complete(_ context.Context, id uuid.UUID) error {
	if f.err != nil {
		return f.err
	}
	f.completed = append(f.completed, id)
	return nil
}

func (f *fakeTodoRepo) Delete(_ context.Context, id uuid.UUID) error {
	if f.err != nil {
		return f.err
	}
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeTodoRepo) List(_ context.Context) ([]todo.Todo, error){
	if f.err != nil {
		return nil, f.err
	}
	return f.created, nil
}

// ── User ──────────────────────────────────────────────────────────────────────

type fakeUserUoW struct {
	repo user.Repository
	err  error
}

func (f *fakeUserUoW) Execute(_ context.Context, fn func(user.Repository) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(f.repo)
}

type fakeUserRepo struct {
	created []user.User
	err     error
}

func (f *fakeUserRepo) Create(_ context.Context, u user.User) error {
	if f.err != nil {
		return f.err
	}
	f.created = append(f.created, u)
	return nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*user.User, error) {
	return nil, f.err
}

// ── Tenant ────────────────────────────────────────────────────────────────────

type fakeTenantUoW struct {
	repo tenant.Repository
	err  error
}

func (f *fakeTenantUoW) Execute(_ context.Context, fn func(tenant.Repository) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(f.repo)
}

type fakeTenantRepo struct {
	created []tenant.Tenant
	err     error
}

func (f *fakeTenantRepo) Create(_ context.Context, t tenant.Tenant) error {
	if f.err != nil {
		return f.err
	}
	f.created = append(f.created, t)
	return nil
}

func (f *fakeTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*tenant.Tenant, error) {
	return nil, f.err
}
