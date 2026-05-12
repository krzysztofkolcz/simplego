Podsumowanie sesji

Co zrobiliśmy

Zaimplementowaliśmy pełny flow ListTodos od OpenAPI spec do HTTP handlera.

OpenAPI spec (apis/mymigrations.yaml) — dodano endpoint GET /todos z operationId listTodos,
parametrem XTenantID w headerze, odpowiedzią 200 ze schematem TodoList. Dodano schemat
TodoList jako type: array z items: $ref Todo.

Konfiguracja codegen (apis/config.yaml) — poprawiono ścieżkę output z ../internal/http/api/api.gen.go
na ./internal/http/api/api.gen.go, żeby komenda działała z roota projektu.

Regeneracja (oapi-codegen) — wygenerowano ListTodos200JSONResponse, ListTodosRequestObject,
ListTodosParams oraz wpis w StrictServerInterface.

Infrastruktura (internal/infrastructure/repository/tenant/todo_repository.go) — zaimplementowano
metodę List() używając r.queryQ.ListTodos(ctx) z mapowaniem querydb.Todo → todo.Todo.

Query handler (internal/application/query/list_todos.go) — stworzono ListTodosQuery (pusty),
TodoResult (ID, Title, Completed), ListTodosHandler z konstruktorem NewListTodosHandler
i metodą Handle zwracającą []TodoResult.

HTTP handler (internal/http/handler/todo_handler.go) — zaimplementowano ListTodos z poprawną
izolacją tenanta przez WithinTransactionReadonly i tenantSchema(req.Params.XTenantID).

Projekt kompiluje się bez błędów (go build ./...).

---
Stan na jutro

Niezcommitowane zmiany — wszystkie pliki z sesji są gotowe do commita.

Znany dług techniczny:
- TodoResult w list_todos.go nie ma pola CreatedAt — HTTP handler przekazuje nil.
  Można dodać CreatedAt do TodoResult i przekazać prawdziwą wartość (SQL ją zwraca).
- Nazwa TodoResult jest zbyt ogólna — konsekwentnie powinna być ListTodoResult.
- GetTodo w todo_handler.go:39 nadal wywołuje WithinTransactionReadonly bezpośrednio
  w handlerze HTTP (naruszenie granic warstw) — ten dług był znany z poprzedniej sesji.

Logiczne kolejne kroki:
1. Commit niezacommitowanych zmian
2. Unit test dla ListTodosHandler — fakeTodoRepo.List już istnieje, test będzie krótki
3. Domain error types zamiast pgx.ErrNoRows w handlerach (dług z poprzedniej sesji)
4. ReadRepository — oddzielny interfejs dla odczytu, żeby handlery query nie wiedziały o TxManager
