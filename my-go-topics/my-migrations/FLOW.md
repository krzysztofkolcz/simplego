# Jak to działa od HTTP do bazy danych

## Warstwy — szybki przegląd

```
HTTP handler          →  odbiera request, tłumaczy na command/query
Application layer     →  logika: command i query handlery
Domain layer          →  interfejsy (Repository, UnitOfWork), encje
Infrastructure layer  →  implementacje: TxManager, Repository, sqlc
Baza danych           →  PostgreSQL z osobnymi schematami per tenant
```

Każda warstwa zna tylko warstwę poniżej przez interfejsy. HTTP handler nie
powinien wiedzieć jak działa transakcja — to jest problem który właśnie
naprawiamy w ReadRepository refaktorze.

---

## Flow 1: Zapis — CreateTodo

Przykład: `POST /todos` z `X-Tenant-ID: abc123`

```
HTTP handler (todo_handler.go)
  │
  │  tworzy: tenantrepo.NewTodoUnitOfWork(s.txManager, "abc123")
  │  tworzy: appcommand.NewCreateTodoHandler(uow)
  │  wywołuje: handler.Handle(ctx, CreateTodoCommand{Title: "Buy milk"})
  │
  ▼
CreateTodoHandler (application/command/create_todo.go)
  │
  │  tworzy encję: todo.Todo{ID: uuid.New(), Title: "Buy milk"}
  │  wywołuje: uow.Execute(ctx, func(repo) { repo.Create(ctx, todo) })
  │
  ▼
TodoUnitOfWork (infrastructure/repository/tenant/unit_of_work.go)
  │
  │  wywołuje: txManager.WithinTransaction(ctx, "abc123", func(commandQ) { ... })
  │  wewnątrz: tworzy TodoRepository(commandQ, queryQ)
  │  wywołuje: fn(repo)  ← to jest func(repo) z Create z handlera wyżej
  │
  ▼
TxManager (infrastructure/db/tx.go)
  │
  │  pool.BeginTx(ctx)           → otwiera transakcję
  │  SET LOCAL search_path="abc123"  → przestawia schemat na tenanta
  │  commanddb.New(tx)           → tworzy Queries opakowane w transakcję
  │  wywołuje: fn(q)             → przekazuje Queries do UoW
  │  tx.Commit()                 → zatwierdza
  │
  ▼
TodoRepository.Create (infrastructure/repository/tenant/todo_repository.go)
  │
  │  wywołuje: commandQ.CreateTodo(ctx, CreateTodoParams{ID, Title})
  │
  ▼
commanddb.Queries.CreateTodo (infrastructure/db/sqlc/command/todos.sql.go)
  │
  │  to jest KOD WYGENEROWANY przez sqlc
  │  wykonuje: INSERT INTO todos (id, title) VALUES ($1, $2) RETURNING ...
  │
  ▼
PostgreSQL — schemat "abc123", tabela todos
```

---

## Flow 2: Odczyt — GetTodo (OBECNY stan — z problemem)

Przykład: `GET /todos/{id}` z `X-Tenant-ID: abc123`

```
HTTP handler (todo_handler.go)
  │
  │  ← PROBLEM: handler sam wywołuje TxManager i zna querydb.Queries
  │
  │  wywołuje: s.txManager.WithinTransactionReadonly(ctx, "abc123", func(q) {
  │      repo := tenantrepo.NewTodoRepository(nil, q)
  │      result, err = appquery.NewGetTodoHandler(repo).Handle(...)
  │  })
  │
  ▼
TxManager.WithinTransactionReadonly (infrastructure/db/tx.go)
  │
  │  pool.BeginTx(ctx, ReadOnly)      → transakcja tylko do odczytu
  │  SET LOCAL search_path="abc123"   → schemat tenanta
  │  querydb.New(tx)                  → Queries opakowane w transakcję
  │  wywołuje: fn(q)
  │
  ▼
GetTodoHandler.Handle (application/query/get_todo.go)
  │
  │  wywołuje: repo.GetByID(ctx, id)
  │
  ▼
TodoRepository.GetByID (infrastructure/repository/tenant/todo_repository.go)
  │
  │  wywołuje: queryQ.GetTodo(ctx, id)
  │  mapuje pgx.ErrNoRows → domain.ErrNotFound
  │
  ▼
querydb.Queries.GetTodo (infrastructure/db/sqlc/query/todos.sql.go)
  │
  │  KOD WYGENEROWANY przez sqlc
  │  wykonuje: SELECT id, title, completed, created_at FROM todos WHERE id=$1
  │
  ▼
PostgreSQL — schemat "abc123", tabela todos
```

---

## Flow 2b: Odczyt — GetTodo (DOCELOWY stan po refaktorze)

```
HTTP handler
  │
  │  tworzy: tenantrepo.NewTodoReadRepository(s.txManager, "abc123")
  │  wywołuje: appquery.NewGetTodoHandler(repo).Handle(...)
  │  ← handler nie wie nic o transakcji
  │
  ▼
GetTodoHandler.Handle (application/query/get_todo.go)
  │
  │  wywołuje: repo.GetByID(ctx, id)
  │  ← repo to ReadRepository z domain — tylko interfejs, nie zna TxManager
  │
  ▼
TodoReadRepository.GetByID (infrastructure/repository/tenant/todo_read_repository.go)
  │
  │  wywołuje: txManager.WithinTransactionReadonly(...)
  │  ← transakcja ukryta w infrastrukturze, nikt wyżej o niej nie wie
  │
  ▼
(dalej tak samo jak wyżej: sqlc → PostgreSQL)
```

---

## Skąd pochodzi sqlc?

sqlc czyta dwa rodzaje plików i generuje Go:

```
db/migrations/tenant/*.sql     → schemat bazy (CREATE TABLE todos ...)
db/queries/todos.sql           → zapytania z adnotacjami:

    -- name: GetTodo :one        ← sqlc generuje func GetTodo() zwracającą jeden wiersz
    SELECT id, title ...
    WHERE id = $1

    -- name: ListTodos :many     ← sqlc generuje func ListTodos() zwracającą slice
    SELECT id, title ...

    -- name: CreateTodo :one     ← wstawia i zwraca wiersz
    INSERT INTO todos ...
```

sqlc generuje dwa osobne pakiety:
- `infrastructure/db/sqlc/command/` — zapytania modyfikujące (INSERT, UPDATE, DELETE)
- `infrastructure/db/sqlc/query/`   — zapytania tylko do odczytu (SELECT)

Podział jest ręczny — sam decydujesz które .sql pliki trafiają do którego pakietu.

---

## Dlaczego dwa zestawy Queries (command i query)?

`commanddb.Queries` i `querydb.Queries` to ten sam wygenerowany typ ale z innego pakietu.
Podział jest architektoniczny, nie techniczny:

- `commanddb` — używany w transakcjach zapisu (WithinTransaction)
- `querydb`   — używany w transakcjach tylko do odczytu (WithinTransactionReadonly)

TxManager otwiera transakcję z odpowiednim trybem i przekazuje właściwy typ Queries.
Repository dostaje ten typ i woła na nim wygenerowane metody SQL.

---

## Dlaczego UnitOfWork zamiast bezpośredniego repozytorium przy zapisie?

UnitOfWork gwarantuje że kilka operacji na repo dzieje się w jednej transakcji:

```go
uow.Execute(ctx, func(repo todo.Repository) error {
    repo.Create(ctx, todo)      // oba w tej samej transakcji
    repo.DoSomethingElse(ctx)   // commit albo rollback razem
    return nil
})
```

Przy odczycie (GetTodo, ListTodos) nie potrzebujemy tej gwarancji — każde zapytanie
może działać niezależnie. Dlatego query handlery używają Repository bezpośrednio,
a nie UnitOfWork.

---

## Podsumowanie — co jest gdzie

| Co                        | Plik                                              |
|---------------------------|---------------------------------------------------|
| SQL dla zapytań           | `db/queries/*.sql`                                |
| Wygenerowany kod (odczyt) | `infrastructure/db/sqlc/query/*.sql.go`           |
| Wygenerowany kod (zapis)  | `infrastructure/db/sqlc/command/*.sql.go`         |
| Otwieranie transakcji     | `infrastructure/db/tx.go` — TxManager             |
| Implementacja repo        | `infrastructure/repository/tenant/todo_repository.go` |
| Interfejsy domain         | `domain/todo/repository.go`, `unit_of_work.go`    |
| UnitOfWork (zapis)        | `infrastructure/repository/tenant/unit_of_work.go`|
| Logika biznesowa          | `application/command/`, `application/query/`      |
| HTTP + routing            | `http/handler/`, `http/router/`                   |
| Wygenerowany HTTP (oapi)  | `http/api/api.gen.go`                             |
