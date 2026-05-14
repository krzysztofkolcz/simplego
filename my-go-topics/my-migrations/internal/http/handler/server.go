package handler

import "github.com/krzysztofkolcz/mymigrations/internal/application/port"

// Server implements httpapi.StrictServerInterface.
// Methods are split across tenant_handler.go, user_handler.go, todo_handler.go.
type Server struct {
	createTodo   port.CreateTodoPort
	completeTodo port.CompleteTodoPort
	deleteTodo   port.DeleteTodoPort
	getTodo      port.GetTodoPort
	listTodos    port.ListTodosPort
	createTenant port.CreateTenantPort
	getTenant    port.GetTenantPort
	createUser   port.CreateUserPort
	getUser      port.GetUserPort
}

func NewServer(
	createTodo port.CreateTodoPort,
	completeTodo port.CompleteTodoPort,
	deleteTodo port.DeleteTodoPort,
	getTodo port.GetTodoPort,
	listTodos port.ListTodosPort,
	createTenant port.CreateTenantPort,
	getTenant port.GetTenantPort,
	createUser port.CreateUserPort,
	getUser port.GetUserPort,
) *Server {
	return &Server{
		createTodo:   createTodo,
		completeTodo: completeTodo,
		deleteTodo:   deleteTodo,
		getTodo:      getTodo,
		listTodos:    listTodos,
		createTenant: createTenant,
		getTenant:    getTenant,
		createUser:   createUser,
		getUser:      getUser,
	}
}
