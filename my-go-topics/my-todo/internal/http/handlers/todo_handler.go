// Package httphandlers is the HTTP inbound adapter.
//
// In Hexagonal Architecture this is an INBOUND ADAPTER: it receives HTTP
// requests and translates them into application-layer Commands and Queries.
//
// TodoHandler depends ONLY on the two buses — it has no direct knowledge of
// repositories, domain aggregates, or database types. The bus is the seam
// that decouples the HTTP layer from the application layer.
//
// Each method maps to one OpenAPI operation:
//   - POST   /todos              → CreateTodoCommand
//   - GET    /todos              → ListTodosQuery
//   - GET    /todos/{id}         → GetTodoByIDQuery
//   - PUT    /todos/{id}         → UpdateTodoCommand
//   - POST   /todos/{id}/complete → CompleteTodoCommand
//   - DELETE /todos/{id}         → DeleteTodoCommand
package httphandlers

import (
	"context"
	"fmt"
	"time"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
	"github.com/C5383717/my-todo/internal/app/bus"
	"github.com/C5383717/my-todo/internal/app/commands"
	"github.com/C5383717/my-todo/internal/app/queries"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TodoHandler implements todoapi.StrictServerInterface.
// It only knows about the two buses — no repositories, no domain types.
type TodoHandler struct {
	cmdBus bus.CommandBus
	qryBus bus.QueryBus
}

func NewTodoHandler(cmdBus bus.CommandBus, qryBus bus.QueryBus) *TodoHandler {
	return &TodoHandler{cmdBus: cmdBus, qryBus: qryBus}
}

// --- Commands (mutate state) ---

// CreateTodo handles POST /todos.
// Generates the UUID here so it can be returned in the 201 response
// without issuing a separate query after the command completes.
func (h *TodoHandler) CreateTodo(
	ctx context.Context,
	req todoapi.CreateTodoRequestObject,
) (todoapi.CreateTodoResponseObject, error) {
	id := uuid.New()

	err := h.cmdBus.Dispatch(ctx, commands.CreateTodoCommand{
		ID:          id,
		Title:       req.Body.Title,
		Description: req.Body.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("create todo: %w", err)
	}

	// After a successful command, run a GetByID query to return the full object.
	// This is the "read-your-writes" pattern — consistent response without
	// maintaining state in the handler.
	result, err := h.qryBus.Ask(ctx, queries.GetTodoByIDQuery{ID: id})
	if err != nil {
		return nil, fmt.Errorf("create todo: read back: %w", err)
	}

	view := result.(queries.TodoView)

	return todoapi.CreateTodo201JSONResponse(toTodoResponse(view)), nil
}

// UpdateTodo handles PUT /todos/{id}.
func (h *TodoHandler) UpdateTodo(
	ctx context.Context,
	req todoapi.UpdateTodoRequestObject,
) (todoapi.UpdateTodoResponseObject, error) {
	id := uuid.UUID(req.Id)

	err := h.cmdBus.Dispatch(ctx, commands.UpdateTodoCommand{
		ID:          id,
		Title:       req.Body.Title,
		Description: req.Body.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}

	result, err := h.qryBus.Ask(ctx, queries.GetTodoByIDQuery{ID: id})
	if err != nil {
		return nil, fmt.Errorf("update todo: read back: %w", err)
	}

	view := result.(queries.TodoView)

	return todoapi.UpdateTodo200JSONResponse(toTodoResponse(view)), nil
}

// CompleteTodo handles POST /todos/{id}/complete.
func (h *TodoHandler) CompleteTodo(
	ctx context.Context,
	req todoapi.CompleteTodoRequestObject,
) (todoapi.CompleteTodoResponseObject, error) {
	id := uuid.UUID(req.Id)

	err := h.cmdBus.Dispatch(ctx, commands.CompleteTodoCommand{ID: id})
	if err != nil {
		return nil, fmt.Errorf("complete todo: %w", err)
	}

	result, err := h.qryBus.Ask(ctx, queries.GetTodoByIDQuery{ID: id})
	if err != nil {
		return nil, fmt.Errorf("complete todo: read back: %w", err)
	}

	view := result.(queries.TodoView)

	return todoapi.CompleteTodo200JSONResponse(toTodoResponse(view)), nil
}

// DeleteTodo handles DELETE /todos/{id}.
// Returns 204 No Content on success — no body needed.
func (h *TodoHandler) DeleteTodo(
	ctx context.Context,
	req todoapi.DeleteTodoRequestObject,
) (todoapi.DeleteTodoResponseObject, error) {
	id := uuid.UUID(req.Id)

	err := h.cmdBus.Dispatch(ctx, commands.DeleteTodoCommand{ID: id})
	if err != nil {
		return nil, fmt.Errorf("delete todo: %w", err)
	}

	return todoapi.DeleteTodo204Response{}, nil
}

// --- Queries (read-only) ---

// GetTodoById handles GET /todos/{id}.
func (h *TodoHandler) GetTodoById(
	ctx context.Context,
	req todoapi.GetTodoByIdRequestObject,
) (todoapi.GetTodoByIdResponseObject, error) {
	result, err := h.qryBus.Ask(ctx, queries.GetTodoByIDQuery{ID: uuid.UUID(req.Id)})
	if err != nil {
		return nil, fmt.Errorf("get todo by id: %w", err)
	}

	view := result.(queries.TodoView)

	return todoapi.GetTodoById200JSONResponse(toTodoResponse(view)), nil
}

// ListTodos handles GET /todos.
func (h *TodoHandler) ListTodos(
	ctx context.Context,
	_ todoapi.ListTodosRequestObject,
) (todoapi.ListTodosResponseObject, error) {
	result, err := h.qryBus.Ask(ctx, queries.ListTodosQuery{})
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}

	viewList := result.([]queries.TodoView)

	items := make([]todoapi.TodoResponse, 0, len(viewList))
	for _, v := range viewList {
		items = append(items, toTodoResponse(v))
	}

	return todoapi.ListTodos200JSONResponse{Items: items}, nil
}

// --- mapping ---

func toTodoResponse(v queries.TodoView) todoapi.TodoResponse {
	return todoapi.TodoResponse{
		Id:          openapi_types.UUID(v.ID),
		Title:       v.Title,
		Description: v.Description,
		Status:      todoapi.TodoResponseStatus(v.Status),
		CreatedAt:   v.CreatedAt.UTC().Truncate(time.Millisecond),
		UpdatedAt:   v.UpdatedAt.UTC().Truncate(time.Millisecond),
	}
}
