package handler

import "github.com/krzysztofkolcz/mymigrations/internal/application/port"

// Server implements httpapi.StrictServerInterface.
// Methods are split across tenant_handler.go, user_handler.go, todo_handler.go, catalog_handler.go.
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

	createComponent        port.CreateComponentPort
	getComponent           port.GetComponentPort
	createProduct          port.CreateProductPort
	getProduct             port.GetProductPort
	setProductComponent    port.SetProductComponentPort
	removeProductComponent port.RemoveProductComponentPort
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
	createComponent port.CreateComponentPort,
	getComponent port.GetComponentPort,
	createProduct port.CreateProductPort,
	getProduct port.GetProductPort,
	setProductComponent port.SetProductComponentPort,
	removeProductComponent port.RemoveProductComponentPort,
) *Server {
	return &Server{
		createTodo:          createTodo,
		completeTodo:        completeTodo,
		deleteTodo:          deleteTodo,
		getTodo:             getTodo,
		listTodos:           listTodos,
		createTenant:        createTenant,
		getTenant:           getTenant,
		createUser:          createUser,
		getUser:             getUser,
		createComponent:     createComponent,
		getComponent:        getComponent,
		createProduct:       createProduct,
		getProduct:          getProduct,
		setProductComponent: setProductComponent,
		removeProductComponent: removeProductComponent,
	}
}
