package handler

import (
	"context"

	"github.com/google/uuid"

	appcommand "github.com/krzysztofkolcz/mymigrations/internal/application/command"
	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
)

// ── Todo ──────────────────────────────────────────────────────────────────────

type fakeCreateTodo struct {
	id  uuid.UUID
	err error
}

func (f *fakeCreateTodo) Handle(_ context.Context, _ appcommand.CreateTodoCommand) (uuid.UUID, error) {
	return f.id, f.err
}

type fakeCompleteTodo struct{ err error }

func (f *fakeCompleteTodo) Handle(_ context.Context, _ appcommand.CompleteTodoCommand) error {
	return f.err
}

type fakeDeleteTodo struct{ err error }

func (f *fakeDeleteTodo) Handle(_ context.Context, _ appcommand.DeleteTodoCommand) error {
	return f.err
}

type fakeGetTodo struct {
	result *appquery.GetTodoResult
	err    error
}

func (f *fakeGetTodo) Handle(_ context.Context, _ appquery.GetTodoQuery) (*appquery.GetTodoResult, error) {
	return f.result, f.err
}

type fakeListTodos struct {
	items []appquery.TodoResult
	err   error
}

func (f *fakeListTodos) Handle(_ context.Context, _ appquery.ListTodosQuery) ([]appquery.TodoResult, error) {
	return f.items, f.err
}

// ── Tenant ────────────────────────────────────────────────────────────────────

type fakeCreateTenant struct {
	id  uuid.UUID
	err error
}

func (f *fakeCreateTenant) Handle(_ context.Context, _ appcommand.CreateTenantCommand) (uuid.UUID, error) {
	return f.id, f.err
}

type fakeGetTenant struct {
	result *appquery.GetTenantResult
	err    error
}

func (f *fakeGetTenant) Handle(_ context.Context, _ appquery.GetTenantQuery) (*appquery.GetTenantResult, error) {
	return f.result, f.err
}

// ── User ──────────────────────────────────────────────────────────────────────

type fakeCreateUser struct {
	id  uuid.UUID
	err error
}

func (f *fakeCreateUser) Handle(_ context.Context, _ appcommand.CreateUserCommand) (uuid.UUID, error) {
	return f.id, f.err
}

type fakeGetUser struct {
	result *appquery.GetUserResult
	err    error
}

func (f *fakeGetUser) Handle(_ context.Context, _ appquery.GetUserQuery) (*appquery.GetUserResult, error) {
	return f.result, f.err
}
