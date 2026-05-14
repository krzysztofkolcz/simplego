# Budowanie aplikacji DDD w Go — krok po kroku

Projekt: aplikacja TODO z multitenancy, ucząca DDD, CQRS, Unit of Work i REST API.
Stack: Go, PostgreSQL, sqlc, golang-migrate, oapi-codegen, testcontainers.

---

## Spis treści

1. [Architektura — warstwy i zasada zależności](#1-architektura)
2. [Warstwa Domain — encje i interfejsy](#2-warstwa-domain)
3. [Warstwa Infrastructure — baza danych](#3-warstwa-infrastructure--baza-danych)
4. [Multitenancy — schema per tenant](#4-multitenancy)
5. [CQRS — podział na command i query](#5-cqrs)
6. [Repository pattern](#6-repository-pattern)
7. [Unit of Work — granica transakcji](#7-unit-of-work)
8. [Warstwa Application — command i query handlery](#8-warstwa-application)
9. [Testy — strategia i wzorce](#9-testy)
10. [Warstwa Presentation — HTTP API](#10-warstwa-presentation--http-api)
11. [OpenAPI i oapi-codegen](#11-openapi-i-oapi-codegen)
12. [Wiring — cmd/api/main.go](#12-wiring)
13. [Przepływ requestu — end to end](#13-przepływ-requestu)
14. [Aggregate z inwariantami](#14-aggregate-z-inwariantami)
15. [Value Objects](#15-value-objects)
16. [Domain Events](#16-domain-events)
17. [ReadRepository — odczyty bez UoW](#17-readrepository)
18. [Input Ports — domknięcie hexagonal architecture](#18-input-ports)
19. [Outbox Publisher — gwarantowana dostawa eventów](#19-outbox-publisher)
20. [Kolejne kroki](#20-kolejne-kroki)

---

## 1. Architektura

Aplikacja zbudowana jest z czterech warstw. Każda warstwa zna tylko warstwy poniżej siebie — nigdy powyżej.

```
cmd/api/main.go              ← punkt wejścia, wiring wszystkich warstw

internal/
  http/                      ← warstwa prezentacji (HTTP)
    handler/                 ← implementacja endpointów
    middleware/              ← logging, recovery, request ID
    router/                  ← konfiguracja serwera HTTP
    api/                     ← wygenerowany kod (oapi-codegen)

  application/               ← orchestracja use case'ów
    command/                 ← operacje mutujące stan
    query/                   ← operacje odczytujące stan
    port/                    ← interfejsy Input Ports (use case interfaces)

  domain/                    ← logika biznesowa, zero zewnętrznych zależności
    todo/
    user/
    tenant/

  infrastructure/            ← szczegóły techniczne
    db/                      ← pool, migracje, TxManager, sqlc
    repository/              ← implementacje interfejsów domenowych
    usecase/                 ← adaptery portów: łączą command/query handlery z infrastrukturą
    event/                   ← implementacje EventPublisher

tests/
  integration/               ← testy z prawdziwą bazą (testcontainers)
```

### Zasada zależności

```
http → port ← usecase (infra)
       ↓            ↓
  application → domain ← repository (infra)
```

- `domain` importuje tylko stdlib, uuid, time — zero zewnętrznych pakietów
- `application` importuje tylko `domain`
- `application/port` importuje `application/command` i `application/query` — definiuje interfejsy use case'ów
- `infrastructure/repository` importuje `domain` i pgx/sqlc
- `infrastructure/usecase` importuje `application/port`, `application/command`, `application/query`, `infrastructure/repository` — łączy warstwy
- `http/handler` importuje tylko `application/port` — zero importów infrastruktury
- `cmd/` importuje wszystko — jedyny punkt wiringu

**Dlaczego to ważne:** możesz podmienić całą bazę danych (np. zmienić z PostgreSQL na MySQL) dotykając tylko `infrastructure/` — `domain`, `application` i `http/handler` zostają bez zmian. HTTP handler nie wie jak use case jest zaimplementowany — zna tylko interfejs.

---

## 2. Warstwa Domain

Domena to serce aplikacji. Nie wie nic o HTTP, bazie danych ani frameworkach.

### Encje

Encja to obiekt z tożsamością (ID). Zawiera dane i — gdy rośnie projekt — logikę biznesową.

```go
// internal/domain/todo/entity.go
type Todo struct {
    ID        uuid.UUID
    Title     string
    Completed bool
    CreatedAt time.Time
}
```

```go
// internal/domain/user/entity.go
type User struct {
    ID    uuid.UUID
    Email Email   // value object — patrz sekcja "Value Objects"
}
```

```go
// internal/domain/tenant/entity.go
type Tenant struct {
    ID               uuid.UUID
    SchemaName       string
    CreatedAt        time.Time
    MigrationStatus  string
    MigrationError   string
    MigrationUpdated time.Time
}
```

### Interfejsy Repository

Domena **definiuje** kontrakt dla dostępu do danych. Nie wie jak jest zaimplementowany.

```go
// internal/domain/todo/repository.go — używany przez command handlery (w transakcji przez UoW)
type Repository interface {
    Create(ctx context.Context, todo Todo) error
    GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
    Update(ctx context.Context, todo Todo) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// internal/domain/todo/read_repository.go — używany przez query handlery (bez UoW)
type ReadRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
    List(ctx context.Context) ([]Todo, error)
}
```

`Repository` i `ReadRepository` to dwa osobne interfejsy: `Repository` działa wewnątrz transakcji (UoW), `ReadRepository` otwiera własną read-only transakcję per-metoda. Patrz sekcja 17.

### Interfejsy Unit of Work

Domena definiuje też kontrakt dla transakcji (wyjaśnienie w rozdziale 7).

```go
// internal/domain/todo/unit_of_work.go
type UnitOfWork interface {
    Execute(ctx context.Context, fn func(repo Repository) error) error
}
```

### Błędy domenowe

Wszystkie błędy domenowe żyją w `internal/domain/errors.go`. Infrastruktura i HTTP mapują je na odpowiednie kody:

```go
var ErrNotFound        = errors.New("not found")
var ErrConflict        = errors.New("conflict")         // 409 — duplikat (naruszenie unique)
var ErrAlreadyCompleted = errors.New("already completed") // inwariant agregatu
var ErrInvalidTitle    = errors.New("title cannot be empty") // inwariant agregatu
var ErrInvalidEmail    = errors.New("invalid email address") // value object
```

HTTP handler mapuje:
- `domain.ErrNotFound` → 404
- `domain.ErrConflict` → 409
- `domain.ErrInvalidTitle`, `domain.ErrInvalidEmail` → 400
- pozostałe → 500

---

## 3. Warstwa Infrastructure — baza danych

### Connection pool

```go
// internal/infrastructure/db/pool.go
pool, err := db.NewPool(ctx, db.Config{
    DatabaseURL:       os.Getenv("DATABASE_URL"),
    MaxConns:          20,
    MinConns:          5,
    MaxConnLifetime:   time.Hour,
    MaxConnIdleTime:   30 * time.Minute,
    HealthCheckPeriod: time.Minute,
}, logger)
```

### Migracje (golang-migrate)

Migracje żyją w `internal/infrastructure/db/migrations/`:

```
migrations/
  public/
    001_init.up.sql    ← tworzy tabele tenants, users
    001_init.down.sql
  tenant/
    001_init.up.sql    ← tworzy tabelę todos (w schemacie tenanta)
    001_init.down.sql
```

Migracje są osadzone w binarce przez `//go:embed`:

```go
//go:embed migrations/*
var MigrationsFS embed.FS
```

Osobne migracje dla public schema i dla każdego tenanta — bo tabela `todos` żyje w schemacie tenanta, a `tenants`/`users` w public.

### sqlc — generowanie kodu Go z SQL

sqlc czyta pliki `.sql` i generuje type-safe Go:

```
db/queries/command/    ← pliki SQL dla INSERT/UPDATE/DELETE
db/queries/query/      ← pliki SQL dla SELECT
db/sqlc/command/       ← wygenerowany Go dla mutacji
db/sqlc/query/         ← wygenerowany Go dla odczytów
```

**Dlaczego dwa osobne zestawy?** To fundament CQRS — wyraźna separacja kodu piszącego i czytającego.

Konfiguracja w `sqlc.yaml` wskazuje które pliki SQL generować do których katalogów.

### accessor.go — rozszerzenie wygenerowanego kodu

sqlc generuje `Queries.db` jako pole nieeksportowane. Potrzebujemy dostępu do niego, żeby przekazać tę samą transakcję do `querydb.Queries`. Dodajemy metodę w osobnym pliku — **nie modyfikujemy** wygenerowanego kodu:

```go
// internal/infrastructure/db/sqlc/command/accessor.go
package commanddb

func (q *Queries) DB() DBTX {
    return q.db
}
```

Dzięki temu możemy w `UnitOfWork` zbudować `queryQ` na tej samej transakcji co `commandQ`:

```go
queryQ := querydb.New(commandQ.DB())
```

---

## 4. Multitenancy

Każdy tenant (klient) ma własny schemat PostgreSQL: `tenant_abc123_def456`.

**Jak to działa:**
1. Przy tworzeniu tenanta: `CREATE SCHEMA "tenant_<uuid>"`
2. Przy każdym zapytaniu: `SET LOCAL search_path = "tenant_<uuid>"`
3. `SET LOCAL` działa tylko w obrębie bieżącej transakcji — bezpieczne w poolingu połączeń

### TxManager

Centralny obiekt zarządzający transakcjami. Ma cztery metody:

```go
// Dla mutacji w schemacie tenanta
func (m *TxManager) WithinTransaction(ctx, tenantSchema string, fn func(*commanddb.Queries) error) error

// Dla odczytów w schemacie tenanta (read-only tx — mniejsze blokowanie)
func (m *TxManager) WithinTransactionReadonly(ctx, tenantSchema string, fn func(*querydb.Queries) error) error

// Dla mutacji w public schema (bez SET search_path)
func (m *TxManager) WithinPublicTransaction(ctx context.Context, fn func(*commanddb.Queries) error) error

// Dla odczytów w public schema
func (m *TxManager) WithinPublicTransactionReadonly(ctx context.Context, fn func(*querydb.Queries) error) error
```

Każda metoda: otwiera transakcję → opcjonalnie ustawia search_path → wykonuje `fn` → commituje lub rollbackuje.

### Schemat nazewnictwa

```go
// internal/infrastructure/db/migrate.go
func CreateSchemaName(tenantID string) string {
    return "tenant_" + strings.ReplaceAll(tenantID, "-", "_")
}
// uuid "abc-123" → "tenant_abc_123"
```

UUID tenanta z nagłówka `X-Tenant-ID` → nazwa schematu → `SET LOCAL search_path`.

---

## 5. CQRS

CQRS (Command Query Responsibility Segregation) rozdziela operacje mutujące stan od odczytujących.

### Na poziomie SQL (sqlc)

```
queries/command/tenant/todos.sql:
  CreateTodo :one
  CompleteTodo :exec
  DeleteTodo :exec

queries/query/tenant/todos.sql:
  GetTodo :one
  ListTodos :many
  ListIncompleteTodos :many
```

### Na poziomie Application layer

```
application/command/    ← CreateTodo, CompleteTodo, DeleteTodo, CreateUser, CreateTenant
application/query/      ← GetTodo, GetUser, GetTenant
```

**Zasada:**
- Command handler: przyjmuje `Command`, zwraca `(uuid.UUID, error)` lub `error`
- Query handler: przyjmuje `Query`, zwraca `(*Result, error)`
- Query handler **nigdy** nie wywołuje `UnitOfWork` — tylko czyta przez `Repository`
- Command handler **nigdy** nie zwraca danych domenowych — tylko ID lub błąd

---

## 6. Repository pattern

### Interfejs w domenie (kontrakt)

```go
// internal/domain/todo/repository.go
type Repository interface {
    Create(ctx context.Context, todo Todo) error
    GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
    Update(ctx context.Context, todo Todo) error   // używane przez CompleteTodo: load → Complete() → Update
    Delete(ctx context.Context, id uuid.UUID) error
}
```

`Complete(id)` zostało usunięte. Wzorzec to teraz: load → mutate → save. Handler ładuje agregat przez `GetByID`, woła `t.Complete()` (metoda na encji sprawdza inwariant), i zapisuje przez `Update`. Patrz sekcja 14.

### Implementacja w infrastrukturze

```go
// internal/infrastructure/repository/tenant/todo_repository.go
type TodoRepository struct {
    commandQ *commanddb.Queries   // dla mutacji
    queryQ   *querydb.Queries     // dla odczytów
}
```

Repozytorium **nie zarządza transakcjami** — dostaje `commandQ`/`queryQ` już skonfigurowane (z właściwą transakcją i search_path) z zewnątrz.

Mapowanie sqlc → domena:

```go
func (r *TodoRepository) GetByID(ctx context.Context, id uuid.UUID) (*todo.Todo, error) {
    row, err := r.queryQ.GetTodo(ctx, id)
    if err != nil {
        return nil, err
    }
    return &todo.Todo{
        ID:        row.ID,
        Title:     row.Title,
        Completed: row.Completed,
        CreatedAt: row.CreatedAt.Time,   // pgtype.Timestamp → time.Time
    }, nil
}
```

### Gdzie żyją repozytoria

```
repository/tenant/     ← operacje w schemacie tenanta (todos)
repository/public/     ← operacje w public schema (tenants, users)
```

---

## 7. Unit of Work

### Problem bez UoW

Bez tego wzorca transakcja musi być otwarta przez **wywołującego** (HTTP handler, test):

```go
// HTTP handler musi znać TxManager i commanddb.Queries — to wyciek infrastruktury
txManager.WithinTransaction(ctx, tenantSchema, func(commandQ *commanddb.Queries) error {
    queryQ := querydb.New(commandQ.DB())
    repo := tenantrepo.NewTodoRepository(commandQ, queryQ)
    handler := command.NewCreateTodoHandler(repo)
    return handler.Handle(...)
})
```

### Rozwiązanie: interfejs w domenie, implementacja w infrastrukturze

```go
// internal/domain/todo/unit_of_work.go — domena wie tylko o kontrakcie
type UnitOfWork interface {
    Execute(ctx context.Context, fn func(repo Repository) error) error
}
```

```go
// internal/infrastructure/repository/tenant/unit_of_work.go — infrastruktura implementuje
type TodoUnitOfWork struct {
    txManager    *db.TxManager
    tenantSchema string
}

func (u *TodoUnitOfWork) Execute(ctx context.Context, fn func(todo.Repository) error) error {
    return u.txManager.WithinTransaction(ctx, u.tenantSchema,
        func(commandQ *commanddb.Queries) error {
            queryQ := querydb.New(commandQ.DB())
            repo := NewTodoRepository(commandQ, queryQ)
            return fn(repo)
        },
    )
}
```

### Command handler — czysty, bez wiedzy o infrastrukturze

```go
type CreateTodoHandler struct {
    uow       todo.UnitOfWork
    publisher domain.EventPublisher   // wstrzykiwany — handler publikuje eventy po transakcji
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
    t, err := todo.NewTodo(uuid.New(), cmd.Title)   // konstruktor waliduje tytuł
    if err != nil {
        return uuid.Nil, err   // domain.ErrInvalidTitle → handler zwraca, HTTP handler mapuje na 400
    }
    err = h.uow.Execute(ctx, func(repo todo.Repository) error {
        return repo.Create(ctx, *t)
    })
    if err != nil {
        return uuid.Nil, err
    }
    // Eventy publikowane po commicie. W produkcji: OutboxPublisher pisze do DB w tej samej tx.
    _ = h.publisher.Publish(ctx, t.PullEvents())
    return t.ID, nil
}
```

### Wiring w HTTP handlerze (per request)

```go
// tenantSchema z nagłówka X-Tenant-ID
uow := tenantrepo.NewTodoUnitOfWork(s.txManager, tenantSchema)
handler := command.NewCreateTodoHandler(uow, s.eventPublisher)
id, err := handler.Handle(ctx, command.CreateTodoCommand{Title: req.Body.Title})
```

### Dla public schema (bez search_path)

```go
// internal/infrastructure/repository/public/tenant_unit_of_work.go
func (u *TenantUnitOfWork) Execute(ctx context.Context, fn func(tenant.Repository) error) error {
    return u.txManager.WithinPublicTransaction(ctx, func(commandQ *commanddb.Queries) error {
        queryQ := querydb.New(commandQ.DB())
        repo := NewTenantRepository(commandQ, queryQ)
        return fn(repo)
    })
}
```

### Gdzie **nie** używać UoW

Query handlery czytają dane — nie potrzebują transakcji z commitem. Dostają `Repository` bezpośrednio:

```go
type GetTodoHandler struct {
    repo todo.Repository   // nie UnitOfWork
}
```

---

## 8. Warstwa Application

### Command handlery

Jeden plik = jeden command. Schemat zawsze taki sam:

```go
// internal/application/command/create_todo.go

type CreateTodoCommand struct {
    TenantSchema string   // schemat PostgreSQL tenanta — wypełniany przez use case adapter
    Title        string
}

type CreateTodoHandler struct {
    uow       todo.UnitOfWork
    publisher domain.EventPublisher
}

func NewCreateTodoHandler(uow todo.UnitOfWork, publisher domain.EventPublisher) *CreateTodoHandler {
    return &CreateTodoHandler{uow: uow, publisher: publisher}
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
    t, err := todo.NewTodo(uuid.New(), cmd.Title)
    if err != nil { return uuid.Nil, err }
    err = h.uow.Execute(ctx, func(repo todo.Repository) error { return repo.Create(ctx, *t) })
    if err != nil { return uuid.Nil, err }
    _ = h.publisher.Publish(ctx, t.PullEvents())
    return t.ID, nil
}
```

`TenantSchema` jest dodane do wszystkich command/query operacji tenant-scoped:
- `CreateTodoCommand`, `CompleteTodoCommand`, `DeleteTodoCommand`
- `GetTodoQuery`, `ListTodosQuery`

Handler aplikacyjny nie używa `TenantSchema` bezpośrednio — dostaje już skonfigurowany `UoW`. Pole jest używane przez use case adapter w warstwie infrastruktury (sekcja 18).

`CompleteTodo` stosuje wzorzec load → mutate → save:

```go
// internal/application/command/complete_todo.go
func (h *CompleteTodoHandler) Handle(ctx context.Context, cmd CompleteTodoCommand) error {
    return h.uow.Execute(ctx, func(repo todo.Repository) error {
        t, err := repo.GetByID(ctx, cmd.ID)
        if err != nil { return err }
        if err := t.Complete(); err != nil { return err }   // sprawdza inwariant
        return repo.Update(ctx, *t)
    })
}
```

### Query handlery

Query handler przyjmuje `ReadRepository` (nie `Repository`), bo nie potrzebuje transakcji z UoW — otwiera własną read-only transakcję:

```go
// internal/application/query/get_todo.go

type GetTodoQuery struct { ID uuid.UUID }

type GetTodoResult struct {
    ID        uuid.UUID
    Title     string
    Completed bool
    CreatedAt time.Time   // zawsze wypełnione — brak null w JSON
}

type GetTodoHandler struct {
    repo todo.ReadRepository   // ReadRepository, nie Repository
}

func (h *GetTodoHandler) Handle(ctx context.Context, q GetTodoQuery) (*GetTodoResult, error) {
    t, err := h.repo.GetByID(ctx, q.ID)
    if err != nil { return nil, err }
    return &GetTodoResult{ID: t.ID, Title: t.Title, Completed: t.Completed, CreatedAt: t.CreatedAt}, nil
}
```

---

## 9. Testy

### Strategia

| Warstwa | Rodzaj testów | Dlaczego |
|---|---|---|
| `domain/` | unit | Czysta logika, zero zależności |
| `application/command/` | unit z fake'ami | Testuj orchestrację bez bazy |
| `application/query/` | unit z fake'ami | j.w. |
| `infrastructure/repository/` | integracyjne | Cała logika to SQL — sens tylko z bazą |
| `http/handler/` | unit z fake portami + httptest | Mapowanie błędów domenowych na kody HTTP — bez bazy |
| Wszystko razem | integracyjne (e2e) | Pewność że warstwy współpracują |

### Unit testy z fake'ami (application layer)

Fake'i żyją w `internal/testhelpers/fake_repo.go` — współdzielone między testami unit i integracyjnymi:

```go
// internal/testhelpers/fake_repo.go

type FakeTodoUoW struct {
    Repo todo.Repository
    Err  error
}
func (f *FakeTodoUoW) Execute(_ context.Context, fn func(todo.Repository) error) error {
    if f.Err != nil { return f.Err }
    return fn(f.Repo)
}

type FakeTodoRepo struct {
    Created []todo.Todo
    Err     error
}
func (f *FakeTodoRepo) Create(_ context.Context, t todo.Todo) error {
    if f.Err != nil { return f.Err }
    f.Created = append(f.Created, t)
    return nil
}
func (f *FakeTodoRepo) GetByID(_ context.Context, id uuid.UUID) (*todo.Todo, error) {
    for i := range f.Created {
        if f.Created[i].ID == id { return &f.Created[i], nil }
    }
    return nil, domain.ErrNotFound
}
func (f *FakeTodoRepo) Update(_ context.Context, t todo.Todo) error {
    for i := range f.Created {
        if f.Created[i].ID == t.ID { f.Created[i] = t; return nil }
    }
    return domain.ErrNotFound
}
// ... Delete, Create

// FakeEventPublisher zbiera eventy do asercji
type FakeEventPublisher struct {
    Published []domain.DomainEvent
    Err       error
}
func (f *FakeEventPublisher) Publish(_ context.Context, events []domain.DomainEvent) error {
    f.Published = append(f.Published, events...)
    return f.Err
}
```

Testy są szybkie (4ms dla 14 testów) i nie wymagają bazy:

```go
func TestCreateTodoHandler_Handle_StoresTodo(t *testing.T) {
    repo := &testhelpers.FakeTodoRepo{}
    pub  := &testhelpers.FakeEventPublisher{}
    handler := NewCreateTodoHandler(&testhelpers.FakeTodoUoW{Repo: repo}, pub)

    id, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

    require.NoError(t, err)
    require.NotEqual(t, uuid.Nil, id)
    require.Equal(t, "buy milk", repo.Created[0].Title)
    require.Len(t, pub.Published, 1)   // TodoCreated event został opublikowany
}
```

Testujemy dwa poziomy błędów osobno:
- `UoWError` — błąd na poziomie transakcji (przed wywołaniem fn)
- `RepoError` — błąd na poziomie operacji na bazie (wewnątrz fn)

### Testy integracyjne (infrastructure + application)

PostgreSQL startuje raz w `TestMain` przez testcontainers. Wszystkie testy dzielą tę samą bazę.

```go
// tests/integration/main_test.go
func TestMain(m *testing.M) {
    container, _ := postgres.Run(ctx, "postgres:15", ...)
    // migracje public
    db.MigratePublic(ctx, TestDsn, logger)
    // tenant testowy
    TenantId = uuid.New().String()
    TenantSchema = db.CreateSchemaName(TenantId)
    db.CreateTenant(ctx, ConnectionPool, TenantId, logger)
    db.MigrateAllTenants(ctx, ConnectionPool, SqlDb, TestDsn, logger)
    m.Run()
}
```

Wzorzec testu repozytorium — dwie transakcje dla izolacji:

```go
func TestTodoRepository_Complete(t *testing.T) {
    // 1. Utwórz (commit)
    txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
        repo := tenantrepo.NewTodoRepository(commandQ, querydb.New(commandQ.DB()))
        return repo.Create(ctx, todo.Todo{ID: id, Title: "test"})
    })

    // 2. Operacja (commit)
    txManager.WithinTransaction(ctx, TenantSchema, func(commandQ *commanddb.Queries) error {
        repo := tenantrepo.NewTodoRepository(commandQ, querydb.New(commandQ.DB()))
        return repo.Complete(ctx, id)
    })

    // 3. Weryfikacja przez readonly tx (po commicie)
    txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
        repo := tenantrepo.NewTodoRepository(nil, queryQ)
        result, _ := repo.GetByID(ctx, id)
        require.True(t, result.Completed)
        return nil
    })
}
```

Wzorzec testu command handlera — handler zarządza transakcją samodzielnie przez UoW:

```go
func TestCreateTodoHandler_Handle(t *testing.T) {
    uow := tenantrepo.NewTodoUnitOfWork(db.NewTxManager(ConnectionPool), TenantSchema)
    handler := command.NewCreateTodoHandler(uow)

    // handler sam otwiera i commituje transakcję przez UoW
    id, err := handler.Handle(ctx, command.CreateTodoCommand{Title: "buy milk"})
    require.NoError(t, err)

    // weryfikacja przez readonly tx
    txManager.WithinTransactionReadonly(ctx, TenantSchema, func(queryQ *querydb.Queries) error {
        repo := tenantrepo.NewTodoRepository(nil, queryQ)
        result, _ := repo.GetByID(ctx, id)
        require.Equal(t, "buy milk", result.Title)
        return nil
    })
}
```

### Unit testy HTTP handlerów (http/handler layer)

Dzięki Input Ports (sekcja 18) `Server` przyjmuje interfejsy — można wstrzyknąć fake bez bazy danych.

**Gdzie żyją fake'i:**

```
internal/http/handler/
  fakes_test.go          ← 9 fake struct implementujących port interfaces (package handler)
  todo_handler_test.go   ← 14 testów
  tenant_handler_test.go ← 5 testów
  user_handler_test.go   ← 5 testów
```

**Wzorzec fake portu** — prosta struct z konfigurowalnymi polami zwracanymi:

```go
// internal/http/handler/fakes_test.go
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
```

**Wzorzec testu** — fake wstrzyknięty bezpośrednio w pole `Server`:

```go
func TestCompleteTodo_AlreadyCompleted_Returns409(t *testing.T) {
    ts := httptest.NewServer(router.New(
        &Server{completeTodo: &fakeCompleteTodo{err: domain.ErrAlreadyCompleted}},
    ))
    defer ts.Close()

    req, _ := http.NewRequest(http.MethodPut,
        ts.URL+"/v1/todos/"+uuid.New().String()+"/complete", nil)
    req.Header.Set("X-Tenant-ID", "00000000-0000-0000-0000-000000000001")

    resp, _ := http.DefaultClient.Do(req)
    assert.Equal(t, http.StatusConflict, resp.StatusCode)
}
```

`&Server{completeTodo: fake}` działa ponieważ testy są w `package handler` — mają dostęp do nieeksportowanych pól. Pozostałe pola `nil` — nie są wywołane przez ten test.

**Pokrycie błędów domenowych:**

| Błąd domenowy | Kod HTTP | Gdzie |
|---|---|---|
| `ErrInvalidTitle` | 400 | CreateTodo |
| `ErrInvalidEmail` | 400 | CreateUser |
| `ErrNotFound` | 404 | GetTodo, CompleteTodo, DeleteTodo, GetTenant, GetUser |
| `ErrAlreadyCompleted` | 409 | CompleteTodo |
| `ErrConflict` | 409 | CreateTenant, CreateUser |
| `errors.New("db down")` | 500 | wszystkie |

25 testów, ~20ms, zero połączeń z bazą danych.

---

## 10. Warstwa Presentation — HTTP API

### Struktura

```
internal/http/
  handler/
    server.go          ← struct Server + NewServer (tylko struct, zero logiki)
    errors.go          ← pomocnicze funkcje: tenantSchema, internalError, notFoundError
    tenant_handler.go  ← CreateTenant, GetTenant
    user_handler.go    ← CreateUser, GetUser
    todo_handler.go    ← CreateTodo, GetTodo, CompleteTodo, DeleteTodo
  middleware/
    requestid.go       ← generuje UUID requestu, wstrzykuje do ctx przez slogctx
    logging.go         ← loguje method/path i status/duration
    recovery.go        ← łapie panic, zwraca JSON 500
  router/
    router.go          ← łączy middleware + strict handler + error handlery
  api/
    api.gen.go         ← wygenerowany przez oapi-codegen (nie edytować)
```

### Server struct

```go
// internal/http/handler/server.go
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
    createTodo   port.CreateTodoPort,
    completeTodo port.CompleteTodoPort,
    deleteTodo   port.DeleteTodoPort,
    getTodo      port.GetTodoPort,
    listTodos    port.ListTodosPort,
    createTenant port.CreateTenantPort,
    getTenant    port.GetTenantPort,
    createUser   port.CreateUserPort,
    getUser      port.GetUserPort,
) *Server { ... }
```

`Server` zna tylko interfejsy portów — zero importów `db`, `tenantrepo`, `publicrepo`. Cała infrastruktura jest schowana za interfejsami. Patrz sekcja 18.

### Metody na rozdzielonych plikach

Go pozwala metodom jednego typu żyć w różnych plikach tego samego pakietu. `Server` implementuje `StrictServerInterface` z oapi-codegen — interfejs z 8 metodami. Metody są rozdzielone po domenach:

```go
// internal/http/handler/todo_handler.go
func (s *Server) CreateTodo(ctx context.Context, req httpapi.CreateTodoRequestObject) (httpapi.CreateTodoResponseObject, error) {
    id, err := s.createTodo.Handle(ctx, appcommand.CreateTodoCommand{
        TenantSchema: tenantSchema(req.Params.XTenantID),
        Title:        req.Body.Title,
    })
    if err != nil {
        if errors.Is(err, domain.ErrInvalidTitle) {
            return httpapi.CreateTodo400JSONResponse{N400JSONResponse: badRequestError(err.Error())}, nil
        }
        return httpapi.CreateTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
    }
    return httpapi.CreateTodo201JSONResponse{Id: id}, nil
}
```

Query handlery analogicznie — tylko wywołanie portu, bez tworzenia repozytoriów:

```go
func (s *Server) GetTodo(ctx context.Context, req httpapi.GetTodoRequestObject) (httpapi.GetTodoResponseObject, error) {
    result, err := s.getTodo.Handle(ctx, appquery.GetTodoQuery{
        TenantSchema: tenantSchema(req.Params.XTenantID),
        ID:           req.Id,
    })
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) { return httpapi.GetTodo404JSONResponse{...}, nil }
        return httpapi.GetTodo500JSONResponse{...}, nil
    }
    return httpapi.GetTodo200JSONResponse{Id: result.ID, Title: result.Title, ...}, nil
}
```

### Middleware chain

```
Incoming request
       ↓
  Recovery       ← łapie panic na każdym poziomie
       ↓
  RequestID      ← generuje UUID, wstrzykuje do ctx
       ↓
  Logging        ← loguje request/response
       ↓
  OAPI handler   ← parsuje parametry, dekoduje body
       ↓
  Handler method ← logika biznesowa
```

Middleware są podane od ostatniego do pierwszego (FILO — First In, Last Out):

```go
Middlewares: []httpapi.MiddlewareFunc{
    middleware.Logging(),    // 3. uruchamia się trzecia
    middleware.RequestID(),  // 2. uruchamia się druga
    middleware.Recovery(),   // 1. uruchamia się pierwsza (outermost)
},
```

---

## 11. OpenAPI i oapi-codegen

### Specyfikacja API

```
apis/
  mymigrations.yaml   ← specyfikacja OpenAPI 3.0.3
  config.yaml         ← konfiguracja oapi-codegen
```

`config.yaml`:
```yaml
package: httpapi
output: ../internal/http/api/api.gen.go
generate:
  models: true
  std-http-server: true
  strict-server: true
  embedded-spec: true
```

- `strict-server: true` generuje `StrictServerInterface` — każda metoda zwraca typed response object zamiast pisać bezpośrednio do `http.ResponseWriter`
- `embedded-spec: true` wbudowuje skompresowany YAML w binarkę — używany przez middleware walidacyjne

### Generowanie kodu

```bash
cd apis && oapi-codegen -config config.yaml mymigrations.yaml
```

### Struktura specyfikacji

```yaml
paths:
  /todos:
    post:
      operationId: createTodo
      parameters:
        - $ref: "#/components/parameters/XTenantID"   # wymagany header
      requestBody:
        $ref: "#/components/schemas/CreateTodoRequest"

components:
  parameters:
    XTenantID:
      name: X-Tenant-ID
      in: header
      required: true
      schema:
        type: string
        format: uuid
```

Nagłówek `X-Tenant-ID` jest wymagany na wszystkich endpointach todos. oapi-codegen generuje automatyczne parsowanie i walidację tego nagłówka.

### Typowane response objects

oapi-codegen generuje osobny typ dla każdego kodu HTTP:

```go
// Zamiast ręcznie pisać do ResponseWriter:
return httpapi.CreateTodo201JSONResponse{Id: id}, nil
return httpapi.CreateTodo404JSONResponse{N404JSONResponse: notFoundError()}, nil
return httpapi.CreateTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
```

Kompilator pilnuje że zwracasz właściwy typ dla danego endpointu.

---

## 12. Wiring

Wszystkie warstwy spinają się w `cmd/api/main.go`:

```go
func main() {
    // 1. Logger (slogctx — propagacja logów przez context)
    slogHandler := slogctx.NewHandler(slog.NewJSONHandler(os.Stdout, nil), nil)
    logger := slog.New(slogHandler)

    // 2. Graceful shutdown — context anulowany przez SIGTERM/SIGINT
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
    defer stop()

    // 3. Database pool
    pool, _ := db.NewPool(ctx, db.Config{DatabaseURL: os.Getenv("DATABASE_URL"), ...}, logger)
    defer pool.Close()

    // 4. Wiring warstw — use case adaptery implementują port interfaces
    txManager     := db.NewTxManager(pool)
    commandQ      := commanddb.New(pool)
    queryQ        := querydb.New(pool)
    eventPublisher := infraevent.NewLogPublisher()

    srv := handler.NewServer(
        usecase.NewCreateTodoUseCase(txManager, eventPublisher),
        usecase.NewCompleteTodoUseCase(txManager),
        usecase.NewDeleteTodoUseCase(txManager),
        usecase.NewGetTodoUseCase(txManager),
        usecase.NewListTodosUseCase(txManager),
        usecase.NewCreateTenantUseCase(txManager),
        usecase.NewGetTenantUseCase(commandQ, queryQ),
        usecase.NewCreateUserUseCase(txManager),
        usecase.NewGetUserUseCase(commandQ, queryQ),
    )
    httpHandler := router.New(srv)   // middleware + routing

    // 5. HTTP server
    httpServer := &http.Server{Addr: ":8080", Handler: httpHandler, ...}
    go httpServer.ListenAndServe()

    // 6. Czekaj na sygnał, graceful shutdown
    <-ctx.Done()
    httpServer.Shutdown(shutdownCtx)
}
```

### Dlaczego use case adaptery są singletonami, choć UoW tworzone per-request?

`UnitOfWork` wymaga `tenantSchema` (specyficznego dla requestu) — ale use case adapter tworzy `UoW` dopiero w metodzie `Handle()`, na podstawie pola `cmd.TenantSchema`. Adapter sam jest bezstanowy i może żyć przez cały czas życia serwera.

`TxManager`, `commandQ`, `queryQ` — singletonem. Use case adaptery — singletonami. `UnitOfWork` — tworzony per-request wewnątrz adaptera.

---

## 13. Przepływ requestu — end to end

Przykład: `POST /v1/todos` z nagłówkiem `X-Tenant-ID: abc-123`

```
1. HTTP request przybywa
       ↓
2. Recovery middleware     — owija handler w defer/recover
       ↓
3. RequestID middleware    — generuje UUID, ctx = slogctx.With(ctx, "requestID", "xyz")
       ↓
4. Logging middleware      — loguje "request received" {method: POST, path: /v1/todos}
       ↓
5. oapi-codegen wrapper    — parsuje X-Tenant-ID header → uuid.UUID
                           — dekoduje JSON body → CreateTodoRequest{Title: "buy milk"}
                           — wywołuje Server.CreateTodo(ctx, CreateTodoRequestObject{...})
       ↓
6. todo_handler.go         — tenantSchema("abc-123") → "tenant_abc_123"
                           — wywołuje s.createTodo.Handle(ctx, CreateTodoCommand{
                               TenantSchema: "tenant_abc_123", Title: "buy milk"
                             })
                           — (createTodo to port.CreateTodoPort — wstrzyknięty przez NewServer)
       ↓
7. createTodoUseCase.go    — uow := tenantrepo.NewTodoUnitOfWork(txManager, "tenant_abc_123")
                           — wywołuje CreateTodoHandler(uow, publisher).Handle(ctx, cmd)
                           — (use case adapter w infrastructure/usecase/ łączy port z infrastrukturą)
       ↓
8. create_todo.go          — t, err := todo.NewTodo(uuid.New(), "buy milk")
                             (NewTodo waliduje tytuł, nagrywa TodoCreated event)
       ↓
9. create_todo.go          — uow.Execute(ctx, fn)
       ↓
10. unit_of_work.go        — txManager.WithinTransaction(ctx, "tenant_abc_123", fn)
       ↓
11. tx.go                  — BEGIN
                           — SET LOCAL search_path = "tenant_abc_123"
                           — commandQ := commanddb.New(tx)
                           — fn(commandQ)
       ↓
12. todo_repository.go     — commandQ.CreateTodo(ctx, CreateTodoParams{ID: uuid, Title: "buy milk"})
       ↓
13. PostgreSQL             — INSERT INTO todos (id, title) VALUES (...)
                             (w schemacie tenant_abc_123)
       ↓
14. tx.go                  — COMMIT
       ↓
15. create_todo.go         — publisher.Publish(ctx, t.PullEvents())
                             (TodoCreated event — po commicie, poza transakcją)
       ↓
16. create_todo.go         — zwraca (id, nil)
       ↓
17. todo_handler.go        — zwraca CreateTodo201JSONResponse{Id: id}
       ↓
18. oapi-codegen           — serializuje do JSON, ustawia status 201
       ↓
19. Logging middleware     — loguje "request completed" {status: 201, duration: 3ms}
       ↓
20. HTTP response 201 {"id": "..."}
```

---

## 14. Aggregate z inwariantami

Agregat to encja z logiką biznesową. Zamiast anemic domain model (settery wszędzie), agregat chroni swoje niezmienniki.

### Konstruktor zamiast pola ID na zewnątrz

```go
// internal/domain/todo/entity.go
type Todo struct {
    ID        uuid.UUID
    Title     string
    Completed bool
    CreatedAt time.Time
    events    []domain.DomainEvent   // prywatne — aggregate records events
}

func NewTodo(id uuid.UUID, title string) (*Todo, error) {
    if strings.TrimSpace(title) == "" {
        return nil, domain.ErrInvalidTitle
    }
    t := &Todo{ID: id, Title: strings.TrimSpace(title)}
    t.record(TodoCreated{TodoID: id, Title: t.Title, OccurredAt: time.Now()})
    return t, nil
}
```

### Metody chronią inwarianty

```go
func (t *Todo) Complete() error {
    if t.Completed {
        return domain.ErrAlreadyCompleted   // nie możesz ukończyć już ukończonego
    }
    t.Completed = true
    t.record(TodoCompleted{TodoID: t.ID, OccurredAt: time.Now()})   // nagrywa event
    return nil
}

func (t *Todo) Delete() {
    t.record(TodoDeleted{TodoID: t.ID, OccurredAt: time.Now()})   // brak inwariantów → brak błędu
}
```

### Wzorzec load → mutate → save

Zamiast `repo.Complete(id)` (baza wie o logice), teraz:

```go
// CompleteTodoHandler
t, err := repo.GetByID(ctx, cmd.ID)   // 1. załaduj agregat
if err != nil { return err }
if err := t.Complete(); err != nil { return err }   // 2. zmutuj (sprawdza inwariant)
return repo.Update(ctx, *t)           // 3. zapisz
```

**Korzyść:** logika jest w kodzie Go (testowalnym), nie rozrzucona po SQL i handlerach.

---

## 15. Value Objects

Value object to typ bez tożsamości — dwa value objects z tymi samymi danymi są równe. Hermetyzuje walidację i normalizację.

### Email

```go
// internal/domain/user/email.go
type Email struct {
    value string   // prywatne — jedyny dostęp przez String()
}

func NewEmail(s string) (Email, error) {
    normalized := strings.ToLower(strings.TrimSpace(s))
    parts := strings.SplitN(normalized, "@", 3)
    if len(parts) != 2 || parts[0] == "" || !strings.Contains(parts[1], ".") {
        return Email{}, domain.ErrInvalidEmail
    }
    return Email{value: normalized}, nil
}

func (e Email) String() string { return e.value }
```

### Użycie w encji

```go
type User struct {
    ID    uuid.UUID
    Email Email   // nie string — kompilator wymusza użycie NewEmail()
}
```

### Propagacja przez warstwy

- **Application**: `email, err := user.NewEmail(cmd.Email)` — walidacja na wejściu command handlera
- **Infrastructure write**: `u.Email.String()` → SQL
- **Infrastructure read**: `email, err := user.NewEmail(row.Email)` → jeśli baza ma zły email, to błąd odczytu (data corruption)
- **HTTP**: `domain.ErrInvalidEmail` → 400

---

## 16. Domain Events

Agregat nagrywa co się stało. Handler publikuje po commicie.

### Interfejsy w domenie

```go
// internal/domain/event.go
type DomainEvent interface {
    EventName() string
}

type EventPublisher interface {
    Publish(ctx context.Context, events []DomainEvent) error
}
```

### Eventy w agregacie

```go
// internal/domain/todo/events.go
type TodoCreated struct {
    TodoID     uuid.UUID
    Title      string
    OccurredAt time.Time
}
func (e TodoCreated) EventName() string { return "todo.created" }

type TodoCompleted struct {
    TodoID     uuid.UUID
    OccurredAt time.Time
}
func (e TodoCompleted) EventName() string { return "todo.completed" }

type TodoDeleted struct {
    TodoID     uuid.UUID
    OccurredAt time.Time
}
func (e TodoDeleted) EventName() string { return "todo.deleted" }
```

Każda mutacja na agregacie nagrywa odpowiedni event:

```go
// entity.go

func NewTodo(id uuid.UUID, title string) (*Todo, error) {
    ...
    t.record(TodoCreated{TodoID: id, Title: t.Title, OccurredAt: time.Now()})
    return t, nil
}

func (t *Todo) Complete() error {
    if t.Completed { return domain.ErrAlreadyCompleted }
    t.Completed = true
    t.record(TodoCompleted{TodoID: t.ID, OccurredAt: time.Now()})
    return nil
}

func (t *Todo) Delete() {
    t.record(TodoDeleted{TodoID: t.ID, OccurredAt: time.Now()})
}

func (t *Todo) PullEvents() []domain.DomainEvent {
    events := t.events
    t.events = nil
    return events
}
```

`Delete()` nie zwraca błędu — nie ma inwariantów do sprawdzenia (todo zawsze można usunąć). `Complete()` zwraca błąd, bo sprawdza inwariant (`ErrAlreadyCompleted`).

### Publikacja po commicie

Wzorzec jest identyczny dla wszystkich trzech handlerów: wykonaj transakcję, a po jej commicie opublikuj eventy z agregatu.

```go
// create_todo.go — t tworzone przed uow.Execute
t, err := todo.NewTodo(uuid.New(), cmd.Title)
err = h.uow.Execute(ctx, func(repo todo.Repository) error { return repo.Create(ctx, *t) })
if err != nil { return uuid.Nil, err }
return t.ID, h.publisher.Publish(ctx, t.PullEvents())

// complete_todo.go — t ładowane wewnątrz closure, ale zadeklarowane na zewnątrz
var t *todo.Todo
err := h.uow.Execute(ctx, func(repo todo.Repository) error {
    var err error
    t, err = repo.GetByID(ctx, cmd.ID)
    if err != nil { return err }
    if err := t.Complete(); err != nil { return err }
    return repo.Update(ctx, *t)
})
if err != nil { return err }
return h.publisher.Publish(ctx, t.PullEvents())

// delete_todo.go — analogicznie: load → Delete() → repo.Delete()
var t *todo.Todo
err := h.uow.Execute(ctx, func(repo todo.Repository) error {
    var err error
    t, err = repo.GetByID(ctx, cmd.ID)
    if err != nil { return err }
    t.Delete()
    return repo.Delete(ctx, cmd.ID)
})
if err != nil { return err }
return h.publisher.Publish(ctx, t.PullEvents())
```

**Dlaczego `DeleteTodoHandler` ładuje agregat?** Żeby agregat mógł nagrać `TodoDeleted`. Bez `GetByID` nie ma obiektu `t`, a eventy żyją na agregacie. Dodatkowa korzyść: handler zwraca `ErrNotFound` na poziomie domeny, nie bazy danych.

**Ważne:** w produkcji `OutboxPublisher` powinien zapisać eventy do tabeli `outbox` **w tej samej transakcji** co agregat (gwarantowana dostawa). `LogPublisher` (aktualna implementacja) tylko loguje — bez gwarancji.

### Implementacje

```go
// internal/infrastructure/event/log_publisher.go — na produkcję (tymczasowo)
type LogPublisher struct{}
func (p *LogPublisher) Publish(ctx context.Context, events []domain.DomainEvent) error {
    for _, e := range events { slog.InfoContext(ctx, "domain event", "event", e.EventName()) }
    return nil
}

// internal/testhelpers/fake_repo.go — do testów
type FakeEventPublisher struct { Published []domain.DomainEvent; Err error }
func (f *FakeEventPublisher) Publish(_ context.Context, events []domain.DomainEvent) error {
    f.Published = append(f.Published, events...)
    return f.Err
}
```

---

## 17. ReadRepository

Dwa osobne interfejsy dla odczytów i zapisów, bo mają inne wymagania transakcyjne.

### Dlaczego nie używać UoW do odczytów?

`UoW.Execute(fn)` otwiera transakcję zapisu — po co blokować zasoby dla SELECT? Query handler potrzebuje tylko jednej operacji per wywołanie.

### ReadRepository: per-metoda transakcja

```go
// internal/domain/todo/read_repository.go
type ReadRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
    List(ctx context.Context) ([]Todo, error)
}
```

```go
// internal/infrastructure/repository/tenant/todo_read_repository.go
type TodoReadRepository struct {
    txManager *db.TxManager
    schema    string
}

func (r *TodoReadRepository) GetByID(ctx context.Context, id uuid.UUID) (*todo.Todo, error) {
    var result *todo.Todo
    err := r.txManager.WithinTransactionReadonly(ctx, r.schema, func(q *querydb.Queries) error {
        repo := NewTodoRepository(nil, q)
        var err error
        result, err = repo.GetByID(ctx, id)
        return err
    })
    return result, err
}
```

Każda metoda chowa `WithinTransactionReadonly` wewnętrznie — wywołujący nie wie nic o transakcjach.

### Porównanie z UoW

| | `Repository` (przez UoW) | `ReadRepository` |
|---|---|---|
| Używane przez | command handlery | query handlery |
| Transakcja | zarządzana przez UoW (może obejmować kilka operacji) | per-metoda, read-only |
| Blokowanie | zapis (może blokować inne tx) | read-only (minimalne blokowanie) |
| Wzorzec | `fn func(repo Repository)` callback | bezpośrednie wywołanie metody |

---

## 18. Input Ports

Input Port to interfejs use case'u — granica między warstwą prezentacji (HTTP) a warstwą aplikacyjną. To ostatni element hexagonal architecture (Ports & Adapters).

### Problem przed refaktorem

HTTP handler tworzył zależności infrastrukturalne per-request:

```go
// PRZED — handler znał infrastrukturę
func (s *Server) CreateTodo(ctx context.Context, req ...) (...) {
    uow := tenantrepo.NewTodoUnitOfWork(s.txManager, tenantSchema(req.Params.XTenantID))
    id, err := appcommand.NewCreateTodoHandler(uow, s.eventPublisher).Handle(...)
}
// Server przechowywał: txManager, queryQ, commandQ, eventPublisher
```

Wyciek: `http/handler` importował `infrastructure/repository/tenant` i `infrastructure/db`.

### Rozwiązanie: Input Port Interface

```go
// internal/application/port/todo.go
type CreateTodoPort interface {
    Handle(ctx context.Context, cmd appcommand.CreateTodoCommand) (uuid.UUID, error)
}

type CompleteTodoPort interface {
    Handle(ctx context.Context, cmd appcommand.CompleteTodoCommand) error
}

type GetTodoPort interface {
    Handle(ctx context.Context, q appquery.GetTodoQuery) (*appquery.GetTodoResult, error)
}

// ... analogicznie: DeleteTodoPort, ListTodosPort, CreateTenantPort, GetTenantPort,
//                   CreateUserPort, GetUserPort
```

### Use Case Adapter — implementacja portu w infrastrukturze

Use case adapter łączy port interface z istniejącymi handlerami aplikacyjnymi:

```go
// internal/infrastructure/usecase/todo_use_case.go

type createTodoUseCase struct {
    txManager      *db.TxManager
    eventPublisher domain.EventPublisher
}

func NewCreateTodoUseCase(txManager *db.TxManager, publisher domain.EventPublisher) port.CreateTodoPort {
    return &createTodoUseCase{txManager: txManager, eventPublisher: publisher}
}

func (u *createTodoUseCase) Handle(ctx context.Context, cmd appcommand.CreateTodoCommand) (uuid.UUID, error) {
    uow := tenantrepo.NewTodoUnitOfWork(u.txManager, cmd.TenantSchema)   // tworzy UoW na podstawie TenantSchema z cmd
    return appcommand.NewCreateTodoHandler(uow, u.eventPublisher).Handle(ctx, cmd)
}
```

Dla query (read-only) analogicznie:

```go
type getTodoUseCase struct {
    txManager *db.TxManager
}

func (u *getTodoUseCase) Handle(ctx context.Context, q appquery.GetTodoQuery) (*appquery.GetTodoResult, error) {
    repo := tenantrepo.NewTodoReadRepository(u.txManager, q.TenantSchema)
    return appquery.NewGetTodoHandler(repo).Handle(ctx, q)
}
```

Adaptery dla operacji public (bez tenanta) nie potrzebują `TenantSchema`:

```go
type getTenantUseCase struct {
    repo tenant.Repository   // repo budowane raz w konstruktorze, bo jest bezstanowe
}

func NewGetTenantUseCase(commandQ *commanddb.Queries, queryQ *querydb.Queries) port.GetTenantPort {
    return &getTenantUseCase{repo: publicrepo.NewTenantRepository(commandQ, queryQ)}
}

func (u *getTenantUseCase) Handle(ctx context.Context, q appquery.GetTenantQuery) (*appquery.GetTenantResult, error) {
    return appquery.NewGetTenantHandler(u.repo).Handle(ctx, q)
}
```

### Przepływ zależności po refaktorze

```
http/handler          application/port           infrastructure/usecase
─────────────         ────────────────           ──────────────────────
Server {              CreateTodoPort ←────────── createTodoUseCase {
  createTodo ──────►   Handle(ctx, cmd)             txManager
}                     }                             eventPublisher
                                                    Handle() {
                                                      uow = NewTodoUnitOfWork(...)
                                                      handler.Handle(ctx, cmd)
                                                    }
```

### Korzyść: unit testowalność HTTP handlerów

Po refaktorze można testować HTTP handler bez bazy danych:

```go
type fakeTodo struct{ returnID uuid.UUID }
func (f *fakeTodo) Handle(_ context.Context, _ appcommand.CreateTodoCommand) (uuid.UUID, error) {
    return f.returnID, nil
}

srv := handler.NewServer(&fakeTodo{returnID: someID}, ...)
ts := httptest.NewServer(router.New(srv))
// test mapowania HTTP — 201, 400, 500 — bez żadnej bazy danych
```

### Gdzie co mieszka

```
application/port/
  todo.go      ← 5 interfejsów: CreateTodoPort, CompleteTodoPort, DeleteTodoPort,
                                 GetTodoPort, ListTodosPort
  tenant.go    ← 2 interfejsy:  CreateTenantPort, GetTenantPort
  user.go      ← 2 interfejsy:  CreateUserPort, GetUserPort

infrastructure/usecase/
  todo_use_case.go    ← 5 adapterów (struct + NewXxx + Handle)
  tenant_use_case.go  ← 2 adaptery
  user_use_case.go    ← 2 adaptery
```

---

## 19. Outbox Publisher

### Problem — utrata eventów między COMMITem a Publish

Aktualny `LogPublisher` publikuje eventy **po** commicie transakcji:

```go
err = h.uow.Execute(ctx, func(repo todo.Repository) error {
    return repo.Create(ctx, *t)   // COMMIT
})
h.publisher.Publish(ctx, t.PullEvents())  // ← co jeśli tu padnie serwer?
```

Między COMMITem a `Publish` może paść serwer, utracić się połączenie, zrestartować kontener. Todo zostaje w bazie, ale `TodoCreated` znika na zawsze. To utrata eventu — naruszenie gwarancji dostawy.

### Rozwiązanie — zapis w tej samej transakcji

Zamiast publishować po commicie, zapisujesz eventy do tabeli `outbox` **w tej samej transakcji** co agregat:

```
BEGIN
  INSERT INTO todos (id, title) ...
  INSERT INTO outbox (event_type, payload, published_at) VALUES ('todo.created', '...', NULL)
COMMIT  ← albo oba rekordy są, albo żaden
```

Osobny worker czyta outbox i publikuje dalej:

```
SELECT * FROM outbox WHERE published_at IS NULL ORDER BY created_at
→ przetwórz event
→ UPDATE outbox SET published_at = NOW()
```

Jeśli serwer padnie po COMMITcie — outbox ma niepublikowane eventy. Worker znajdzie je przy restarcie. Gwarancja **at-least-once delivery** (event może dotrzeć więcej niż raz — konsument powinien być idempotentny).

### Czy potrzebujesz brokera wiadomości?

Nie. Outbox Pattern to wzorzec niezależny od docelowego miejsca eventu:

| Cel publikacji | Kiedy |
|---|---|
| Handler w tym samym procesie | Najprostszy przypadek — wystarczy Go + PostgreSQL |
| Zewnętrzny broker (Kafka, NATS, RabbitMQ) | Gdy inne serwisy muszą reagować na eventy |
| Webhook / HTTP call | Integracja zewnętrznych systemów |

### Narzędzia (od prostych do złożonych)

**Go + PostgreSQL (ten projekt)** — tabela `outbox`, goroutine co kilka sekund. Zero nowej infrastruktury. Wystarczy do nauki wzorca i produkcji małej skali.

**River** (`riverqueue/river`) — biblioteka Go do job queue na PostgreSQL. Implementuje Outbox Pattern wewnętrznie, obsługuje retry, dead-letter queue, scheduling. Najlepszy wybór gdy chcesz "outbox gotowy z pudełka" na PostgreSQL.

**NATS** — lekki broker w Go, bardzo prosty. Używany jako *cel* eventu z outboxa (worker → NATS → subskrybenci). Nie rozwiązuje samego outboxa.

**RabbitMQ** — klasyczny broker AMQP. Bardziej rozbudowany niż NATS, lepszy dla złożonych topologii routingu. Zewnętrzna infrastruktura.

**Kafka** — distributed log, nie queue. Eventy trwałe i odtwarzalne. Używany przy Event Sourcing, dużej skali, wielu konsumentach. Ciężka infrastruktura — overkill dla małego projektu.

### Implementacja w tym projekcie

```
infrastructure/event/
  log_publisher.go     ← poprzednia implementacja (tylko logi)
  outbox_publisher.go  ← nowa: zapisuje do tabeli outbox w transakcji

infrastructure/db/
  migrations/public/
    002_outbox.up.sql  ← CREATE TABLE outbox_events (...)
  txctx.go            ← context key: ContextWithTxQueries / TxQueriesFromCtx

infrastructure/worker/
  outbox_worker.go    ← goroutine co 5s: czyta outbox → log → mark published
```

**Tabela outbox:**

```sql
CREATE TABLE outbox_events (
    id           BIGSERIAL PRIMARY KEY,
    event_name   TEXT NOT NULL,
    payload      JSONB NOT NULL,
    created_at   TIMESTAMP DEFAULT now(),
    published_at TIMESTAMP
);
```

### Kluczowa zmiana — UnitOfWork.Execute fn dostaje ctx

`domain.EventPublisher` dostaje tylko `context.Context` — nie `*commanddb.Queries`. Żeby `OutboxPublisher` pisał do bazy w tej samej transakcji, queries muszą płynąć przez context.

**Rozwiązanie:** zmiana sygnatury `fn` w `todo.UnitOfWork`:

```go
// przed:
Execute(ctx context.Context, fn func(repo Repository) error) error

// po:
Execute(ctx context.Context, fn func(ctx context.Context, repo Repository) error) error
```

`TodoUnitOfWork.Execute` tworzy `txCtx` z osadzonymi queries i przekazuje go do `fn`:

```go
func (u *TodoUnitOfWork) Execute(ctx context.Context, fn func(context.Context, todo.Repository) error) error {
    return u.txManager.WithinTransaction(ctx, u.tenantSchema, func(commandQ *commanddb.Queries) error {
        txCtx := db.ContextWithTxQueries(ctx, commandQ)
        repo := NewTodoRepository(commandQ, querydb.New(commandQ.DB()))
        return fn(txCtx, repo)
    })
}
```

Command handlery wołają `publisher.Publish` **wewnątrz** `fn` (czyli w transakcji):

```go
func (h *CompleteTodoHandler) Handle(ctx context.Context, cmd CompleteTodoCommand) error {
    return h.uow.Execute(ctx, func(txCtx context.Context, repo todo.Repository) error {
        t, err := repo.GetByID(txCtx, cmd.ID)
        if err != nil { return err }
        if err := t.Complete(); err != nil { return err }
        if err := repo.Update(txCtx, *t); err != nil { return err }
        return h.publisher.Publish(txCtx, t.PullEvents())  // ← wewnątrz transakcji
    })
}
```

**OutboxPublisher** wyciąga queries z context i wstawia event:

```go
func (p *OutboxPublisher) Publish(ctx context.Context, events []domain.DomainEvent) error {
    q, ok := db.TxQueriesFromCtx(ctx)
    if !ok {
        return nil  // brak transakcji (np. fake UoW w testach) → no-op
    }
    for _, e := range events {
        payload, _ := json.Marshal(e)
        q.InsertOutboxEvent(ctx, commanddb.InsertOutboxEventParams{
            EventName: e.EventName(),
            Payload:   payload,
        })
    }
    return nil
}
```

**OutboxWorker** — goroutine w main.go:

```go
func (w *OutboxWorker) processOnce(ctx context.Context) error {
    events, _ := w.queries.SelectUnpublishedOutboxEvents(ctx)
    for _, e := range events {
        slog.InfoContext(ctx, "outbox: dispatching event", "id", e.ID, "event_name", e.EventName)
        w.queries.MarkOutboxEventPublished(ctx, e.ID)
    }
    return nil
}
```

**Wiring w main.go:**

```go
eventPublisher := infraevent.NewOutboxPublisher()
outboxWorker := worker.NewOutboxWorker(commanddb.New(pool), 5*time.Second)
go outboxWorker.Run(ctx)
```

### Gwarancje i właściwości

- **Atomowość**: agregat + event w outbox commitują się razem lub w ogóle
- **At-least-once**: jeśli worker padnie po mark-published, event nie jest ponownie wysyłany; jeśli przed — zostanie wysłany jeszcze raz (konsument powinien być idempotentny)
- **Kompatybilność z testami**: `FakeTodoUoW` wywołuje `fn(context.Background(), repo)` — `OutboxPublisher` dostaje pusty context, `TxQueriesFromCtx` zwraca `false`, zwraca `nil` — żadnego efektu

---

## 20. Kolejne kroki

### Zrobione (sesje 2026-05-12, 2026-05-13, 2026-05-14)
- [x] `ListTodos` query handler + ReadRepository
- [x] `ErrConflict` (409) dla CreateTenant i CreateUser
- [x] `CreatedAt` w GetTodoResult i ListTodosResult
- [x] Testy integracyjne HTTP (httptest.NewServer + router.New)
- [x] Unit testy query handlerów
- [x] Aggregate z inwariantami (`NewTodo`, `Complete()`)
- [x] Value Object `Email`
- [x] Domain Events (`TodoCreated`, `LogPublisher`, `FakeEventPublisher`)
- [x] Input Ports — `application/port/` + `infrastructure/usecase/` — domknięcie hexagonal architecture
- [x] `ErrAlreadyCompleted` → 409 w OpenAPI spec i handlerze
- [x] Domain Events `TodoCompleted` i `TodoDeleted` — `Complete()` i `Delete()` nagrywają eventy, handlery publikują po commicie
- [x] Unit testy HTTP handlerów — 25 testów z fake portami, ~20ms, zero bazy danych

### Do zrobienia
- [ ] `OutboxPublisher` — zapis eventów do tabeli DB w tej samej transakcji (gwarantowana dostawa)
- [ ] Middleware autoryzacji — JWT, wyciąganie tenant ID z tokena zamiast z nagłówka

### Diagram docelowej architektury z Event Sourcing

```
HTTP Request
    ↓
Command Handler
    ↓
Domain Entity (generuje Event)
    ↓
Event Store (ta sama transakcja)     Read Model (projekcja)
    ↓                                      ↑
Event Bus ─────────────────────────────────┘
                                     Query Handler
                                           ↓
                                     HTTP Response
```
