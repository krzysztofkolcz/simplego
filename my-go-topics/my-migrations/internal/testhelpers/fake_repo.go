package testhelpers

import (
	"context"

	"github.com/google/uuid"

	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/tenant"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/todo"
	"github.com/krzysztofkolcz/mymigrations/internal/domain/user"
)

// ── EventPublisher ────────────────────────────────────────────────────────────

type FakeEventPublisher struct {
	Published []domain.DomainEvent
	Err       error
}

func (p *FakeEventPublisher) Publish(_ context.Context, events []domain.DomainEvent) error {
	if p.Err != nil {
		return p.Err
	}
	p.Published = append(p.Published, events...)
	return nil
}

// ── Todo ──────────────────────────────────────────────────────────────────────

type FakeTodoUoW struct {
	Repo todo.Repository
	Err  error // returned by Execute before calling fn
}

func (f *FakeTodoUoW) Execute(_ context.Context, fn func(context.Context, todo.Repository) error) error {
	if f.Err != nil {
		return f.Err
	}
	return fn(context.Background(), f.Repo)
}

type FakeTodoRepo struct {
	Created []todo.Todo
	Deleted []uuid.UUID
	Err     error // returned by every method
}

func (f *FakeTodoRepo) Create(_ context.Context, t todo.Todo) error {
	if f.Err != nil {
		return f.Err
	}
	f.Created = append(f.Created, t)
	return nil
}

func (f *FakeTodoRepo) GetByID(_ context.Context, id uuid.UUID) (*todo.Todo, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	for i, t := range f.Created {
		if t.ID == id {
			return &f.Created[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (f *FakeTodoRepo) Update(_ context.Context, t todo.Todo) error {
	if f.Err != nil {
		return f.Err
	}
	for i, existing := range f.Created {
		if existing.ID == t.ID {
			f.Created[i] = t
			return nil
		}
	}
	return domain.ErrNotFound
}

func (f *FakeTodoRepo) Delete(_ context.Context, id uuid.UUID) error {
	if f.Err != nil {
		return f.Err
	}
	f.Deleted = append(f.Deleted, id)
	return nil
}

func (f *FakeTodoRepo) List(_ context.Context) ([]todo.Todo, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Created, nil
}

// ── User ──────────────────────────────────────────────────────────────────────

type FakeUserUoW struct {
	Repo user.Repository
	Err  error
}

func (f *FakeUserUoW) Execute(_ context.Context, fn func(user.Repository) error) error {
	if f.Err != nil {
		return f.Err
	}
	return fn(f.Repo)
}

type FakeUserRepo struct {
	Created []user.User
	Err     error
}

func (f *FakeUserRepo) Create(_ context.Context, u user.User) error {
	if f.Err != nil {
		return f.Err
	}
	f.Created = append(f.Created, u)
	return nil
}

func (f *FakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	for i, u := range f.Created {
		if u.ID == id {
			return &f.Created[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

// ── Tenant ────────────────────────────────────────────────────────────────────

type FakeTenantUoW struct {
	Repo tenant.Repository
	Err  error
}

func (f *FakeTenantUoW) Execute(_ context.Context, fn func(tenant.Repository) error) error {
	if f.Err != nil {
		return f.Err
	}
	return fn(f.Repo)
}

type FakeTenantRepo struct {
	Created []tenant.Tenant
	Err     error
}

func (f *FakeTenantRepo) Create(_ context.Context, t tenant.Tenant) error {
	if f.Err != nil {
		return f.Err
	}
	f.Created = append(f.Created, t)
	return nil
}

func (f *FakeTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*tenant.Tenant, error) {
	return nil, f.Err
}