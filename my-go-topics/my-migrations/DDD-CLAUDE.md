# DDD — notatki z sesji

Projekt: prosta aplikacja TODO z multitenancy, ucząca DDD, CQRS i Unit of Work.
Stack: Go, PostgreSQL, sqlc, golang-migrate, testcontainers.

---

## Struktura warstw

```
internal/
  domain/                        ← czysta logika biznesowa, zero zewnętrznych zależności
    todo/
      entity.go                  ← encja Todo (ID, Title, Completed, CreatedAt)
      repository.go              ← interfejs Repository
      unit_of_work.go            ← interfejs UnitOfWork
    tenant/
      entity.go
      repository.go

  application/                   ← orchestracja use case'ów, zna tylko interfejsy domenowe
    command/
      create_todo.go
      complete_todo.go
      delete_todo.go
    query/
      get_todo.go

  infrastructure/                ← szczegóły techniczne, implementuje interfejsy domenowe
    db/
      pool.go                    ← pgxpool
      migrate.go                 ← golang-migrate
      tx.go                      ← TxManager
      sqlc/
        command/                 ← wygenerowany sqlc dla mutacji
        query/                   ← wygenerowany sqlc dla odczytów
    repository/
      tenant/
        todo_repository.go       ← implementacja todo.Repository
        unit_of_work.go          ← implementacja todo.UnitOfWork
      public/
        tenant_repository.go     ← implementacja tenant.Repository

tests/
  integration/                   ← testy z prawdziwą bazą (testcontainers)
```

### Zasada zależności

Zależności płyną tylko do wewnątrz:

```
infrastructure → application → domain
```

- `domain` nie importuje niczego spoza biblioteki standardowej i uuid/time
- `application` importuje tylko `domain`
- `infrastructure` importuje `domain` i zewnętrzne biblioteki (pgx, sqlc)
- `cmd/` i testy importują wszystko — to punkt wiring

---

## Repository pattern

### Interfejs w domenie

```go
// internal/domain/todo/repository.go
type Repository interface {
    Create(ctx context.Context, todo Todo) error
    GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
    Complete(ctx context.Context, id uuid.UUID) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

Interfejs żyje w domenie. Application layer zna tylko ten interfejs — nigdy konkretną implementację.

### Implementacja w infrastrukturze

```go
// internal/infrastructure/repository/tenant/todo_repository.go
type TodoRepository struct {
    commandQ *commanddb.Queries   // dla mutacji (INSERT, UPDATE, DELETE)
    queryQ   *querydb.Queries     // dla odczytów (SELECT)
}
```

`commandQ` i `queryQ` są wstrzykiwane z zewnątrz — repozytorium nie zarządza transakcjami.

---

## CQRS — podział na command i query

sqlc generuje osobny kod dla mutacji i odczytów:

```
db/queries/command/   ← SQL dla INSERT/UPDATE/DELETE
db/queries/query/     ← SQL dla SELECT
db/sqlc/command/      ← wygenerowany Go dla mutacji
db/sqlc/query/        ← wygenerowany Go dla odczytów
```

W application layer:

```
application/command/  ← handlery mutujące stan (Create, Complete, Delete)
application/query/    ← handlery odczytujące stan (Get)
```

Command handler zwraca tylko błąd (lub ID nowo utworzonego zasobu).
Query handler zwraca dane — nigdy nie mutuje stanu.

---

## Multitenancy — schema per tenant

Każdy tenant ma własny schemat PostgreSQL (np. `tenant_abc123`).
Izolacja osiągana przez `SET LOCAL search_path = "tenant_abc123"` na początku transakcji.

`TxManager` obsługuje to transparentnie:

```go
// internal/infrastructure/db/tx.go

// Dla mutacji — otwiera transakcję read-write, ustawia search_path
func (m *TxManager) WithinTransaction(ctx, tenantSchema, fn func(*commanddb.Queries) error) error

// Dla odczytów — otwiera transakcję read-only, ustawia search_path
func (m *TxManager) WithinTransactionReadonly(ctx, tenantSchema, fn func(*querydb.Queries) error) error
```

`SET LOCAL` — search_path obowiązuje tylko w obrębie transakcji, po jej zakończeniu wraca do domyślnego.

---

## Unit of Work

### Problem bez UoW

Bez Unit of Work wywołujący (test, HTTP handler) musiał:
1. Otworzyć transakcję
2. Stworzyć repo
3. Wstrzyknąć repo do handlera
4. Wywołać handler
5. Zacommitować

To oznacza, że szczegóły infrastruktury (transakcje, sqlc) wyciekają do warstwy wywołującej.

### Rozwiązanie — interfejs w domenie

```go
// internal/domain/todo/unit_of_work.go
type UnitOfWork interface {
    Execute(ctx context.Context, fn func(repo Repository) error) error
}
```

### Implementacja w infrastrukturze

```go
// internal/infrastructure/repository/tenant/unit_of_work.go
type TodoUnitOfWork struct {
    txManager    *db.TxManager
    tenantSchema string
}

func (u *TodoUnitOfWork) Execute(ctx context.Context, fn func(todo.Repository) error) error {
    return u.txManager.WithinTransaction(ctx, u.tenantSchema, func(commandQ *commanddb.Queries) error {
        queryQ := querydb.New(commandQ.DB())
        repo := NewTodoRepository(commandQ, queryQ)
        return fn(repo)
    })
}
```

`commandQ.DB()` — metoda dodana ręcznie w `accessor.go` obok wygenerowanego kodu sqlc,
zwraca `DBTX` (interfejs pgx), dzięki czemu `queryQ` operuje na tej samej transakcji co `commandQ`.

### Command handler po refaktorze

```go
type CreateTodoHandler struct {
    uow todo.UnitOfWork   // zna tylko interfejs domenowy
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
    t := todo.Todo{ID: uuid.New(), Title: cmd.Title}
    err := h.uow.Execute(ctx, func(repo todo.Repository) error {
        return repo.Create(ctx, t)
    })
    ...
}
```

Handler nie wie nic o transakcjach, pgx, sqlc ani schematach.

### Wiring (cmd/api lub test)

```go
uow     := tenantrepo.NewTodoUnitOfWork(txManager, tenantSchema)
handler := command.NewCreateTodoHandler(uow)
id, err := handler.Handle(ctx, command.CreateTodoCommand{Title: "buy milk"})
```

---

## Gdzie używać WithinTransaction i WithinTransactionReadonly

| Miejsce                          | Metoda                        | Dlaczego                                      |
|----------------------------------|-------------------------------|-----------------------------------------------|
| `TodoUnitOfWork.Execute`         | `WithinTransaction`           | mutacje wymagają read-write tx z search_path  |
| `query handler` / wiring         | `WithinTransactionReadonly`   | odczyty — mniejsze blokowanie, bez zapisu     |
| testy integracyjne (weryfikacja) | `WithinTransactionReadonly`   | sprawdzenie stanu po commicie command handlera|
| testy repozytorium               | obie, osobno dla każdej fazy  | write → commit → read                        |

**Nigdy** nie wywołuj `WithinTransaction` bezpośrednio w command handlerze ani w HTTP handlerze.
Rób to przez `UnitOfWork` — inaczej infrastruktura wycieka do warstwy aplikacji.

---

## Testy integracyjne

Używają testcontainers — PostgreSQL startuje raz w `TestMain`, wszyscy testy dzielą tę samą bazę.

```
tests/integration/
  main_test.go                   ← TestMain: kontener, migracje, tenant testowy
  migration_test.go              ← testy migracji publicznych i tenantowych
  todo_repository_test.go        ← testy warstwy infrastruktury (repo)
  todo_command_handler_test.go   ← testy warstwy aplikacji (command handlery)
```

### Wzorzec test repozytorium

```
WithinTransaction    → Create/Complete/Delete (commit)
WithinTransactionReadonly → GetByID (weryfikacja po commicie)
```

Dwie osobne transakcje — gwarantuje, że test sprawdza rzeczywistą persystencję, nie tylko stan w pamięci transakcji.

### Wzorzec test command handlera

```go
uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
handler := command.NewCreateTodoHandler(uow)

id, err := handler.Handle(ctx, command.CreateTodoCommand{Title: "buy milk"})
require.NoError(t, err)

// weryfikacja przez warstwę query
txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error { ... })
```

---

## Kolejne kroki (do zrobienia)

- [ ] Query handler (`GetTodoHandler`) z własnym mechanizmem transakcji readonly
- [ ] HTTP handler (Chi) — wiring tenantSchema z JWT/kontekstu requestu
- [ ] CQRS pełne — osobna baza / projekcja dla read modelu
- [ ] Event Sourcing — command handler emituje `DomainEvent`, read model jest projekcją zdarzeń
