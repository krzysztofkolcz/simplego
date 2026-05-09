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
14. [Kolejne kroki](#14-kolejne-kroki)

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

  domain/                    ← logika biznesowa, zero zewnętrznych zależności
    todo/
    user/
    tenant/

  infrastructure/            ← szczegóły techniczne
    db/                      ← pool, migracje, TxManager, sqlc
    repository/              ← implementacje interfejsów domenowych

tests/
  integration/               ← testy z prawdziwą bazą (testcontainers)
```

### Zasada zależności

```
http → application → domain ← infrastructure
```

- `domain` importuje tylko stdlib, uuid, time — zero zewnętrznych pakietów
- `application` importuje tylko `domain`
- `infrastructure` importuje `domain` i zewnętrzne biblioteki (pgx, sqlc)
- `http` importuje `application`, `domain`, `infrastructure`
- `cmd/` importuje wszystko — to jedyny punkt wiring

**Dlaczego to ważne:** możesz podmienić całą bazę danych (np. zmienić z PostgreSQL na MySQL) dotykając tylko `infrastructure/` — `domain` i `application` zostają bez zmian.

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
    Email string
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
// internal/domain/todo/repository.go
type Repository interface {
    Create(ctx context.Context, todo Todo) error
    GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
    Complete(ctx context.Context, id uuid.UUID) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Interfejsy Unit of Work

Domena definiuje też kontrakt dla transakcji (wyjaśnienie w rozdziale 7).

```go
// internal/domain/todo/unit_of_work.go
type UnitOfWork interface {
    Execute(ctx context.Context, fn func(repo Repository) error) error
}
```

### Kiedy dodać logikę do encji?

Gdy pojawi się reguła biznesowa, przenieś ją do encji jako konstruktor lub metodę:

```go
// Przykład: walidacja przy tworzeniu
func NewTodo(title string) (Todo, error) {
    if strings.TrimSpace(title) == "" {
        return Todo{}, errors.New("title cannot be empty")
    }
    if len(title) > 255 {
        return Todo{}, errors.New("title too long")
    }
    return Todo{ID: uuid.New(), Title: title}, nil
}
```

Taka logika jest łatwa do przetestowania — żadnych zależności zewnętrznych.

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
    Complete(ctx context.Context, id uuid.UUID) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

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
    uow todo.UnitOfWork   // tylko interfejs domenowy
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
    t := todo.Todo{ID: uuid.New(), Title: cmd.Title}
    err := h.uow.Execute(ctx, func(repo todo.Repository) error {
        return repo.Create(ctx, t)
    })
    if err != nil {
        return uuid.Nil, err
    }
    return t.ID, nil
}
```

### Wiring w HTTP handlerze (per request)

```go
// tenantSchema z nagłówka X-Tenant-ID
uow := tenantrepo.NewTodoUnitOfWork(s.txManager, tenantSchema)
handler := command.NewCreateTodoHandler(uow)
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
    Title string
}

type CreateTodoHandler struct {
    uow todo.UnitOfWork
}

func NewCreateTodoHandler(uow todo.UnitOfWork) *CreateTodoHandler {
    return &CreateTodoHandler{uow: uow}
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (uuid.UUID, error) {
    t := todo.Todo{ID: uuid.New(), Title: cmd.Title}
    err := h.uow.Execute(ctx, func(repo todo.Repository) error {
        return repo.Create(ctx, t)
    })
    if err != nil {
        return uuid.Nil, err
    }
    return t.ID, nil
}
```

Handlery dla `CompleteTodo` i `DeleteTodo` są analogiczne ale zwracają tylko `error`.

### Query handlery

```go
// internal/application/query/get_todo.go

type GetTodoQuery struct { ID uuid.UUID }

type GetTodoResult struct {
    ID        uuid.UUID
    Title     string
    Completed bool
}

type GetTodoHandler struct {
    repo todo.Repository
}

func (h *GetTodoHandler) Handle(ctx context.Context, q GetTodoQuery) (*GetTodoResult, error) {
    t, err := h.repo.GetByID(ctx, q.ID)
    if err != nil {
        return nil, err
    }
    return &GetTodoResult{ID: t.ID, Title: t.Title, Completed: t.Completed}, nil
}
```

`GetTodoResult` celowo nie zawiera wszystkich pól encji — zwraca tylko to, czego potrzebuje wywołujący.

---

## 9. Testy

### Strategia

| Warstwa | Rodzaj testów | Dlaczego |
|---|---|---|
| `domain/` | unit | Czysta logika, zero zależności |
| `application/command/` | unit z fake'ami | Testuj orchestrację bez bazy |
| `application/query/` | unit z fake'ami | j.w. |
| `infrastructure/repository/` | integracyjne | Cała logika to SQL — sens tylko z bazą |
| `http/handler/` | unit / httptest | Mapowanie błędów na kody HTTP |
| Wszystko razem | integracyjne (e2e) | Pewność że warstwy współpracują |

### Unit testy z fake'ami (application layer)

Fake UoW i fake Repository implementują interfejsy domenowe w pamięci:

```go
// internal/application/command/fakes_test.go

type fakeTodoUoW struct {
    repo todo.Repository
    err  error
}

func (f *fakeTodoUoW) Execute(_ context.Context, fn func(todo.Repository) error) error {
    if f.err != nil { return f.err }
    return fn(f.repo)
}

type fakeTodoRepo struct {
    created   []todo.Todo
    completed []uuid.UUID
    err       error
}

func (f *fakeTodoRepo) Create(_ context.Context, t todo.Todo) error {
    if f.err != nil { return f.err }
    f.created = append(f.created, t)
    return nil
}
// ... pozostałe metody
```

Testy są szybkie (4ms dla 14 testów) i nie wymagają bazy:

```go
func TestCreateTodoHandler_Handle_StoresTodo(t *testing.T) {
    repo := &fakeTodoRepo{}
    handler := NewCreateTodoHandler(&fakeTodoUoW{repo: repo})

    id, err := handler.Handle(context.Background(), CreateTodoCommand{Title: "buy milk"})

    require.NoError(t, err)
    require.NotEqual(t, uuid.Nil, id)
    require.Equal(t, "buy milk", repo.created[0].Title)
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
    txManager *db.TxManager
    queryQ    *querydb.Queries    // pool-backed, dla public schema reads
    commandQ  *commanddb.Queries  // pool-backed, dla public schema reads
}

func NewServer(txManager *db.TxManager, pool *pgxpool.Pool) *Server {
    return &Server{
        txManager: txManager,
        queryQ:    querydb.New(pool),
        commandQ:  commanddb.New(pool),
    }
}
```

Queries dla public schema (tenants, users) używają pola bezpośrednio z pool — nie potrzebują transakcji z SET search_path, bo public jest domyślnym schematem.

### Metody na rozdzielonych plikach

Go pozwala metodom jednego typu żyć w różnych plikach tego samego pakietu. `Server` implementuje `StrictServerInterface` z oapi-codegen — interfejs z 8 metodami. Metody są rozdzielone po domenach:

```go
// internal/http/handler/todo_handler.go
func (s *Server) CreateTodo(ctx context.Context, req httpapi.CreateTodoRequestObject) (httpapi.CreateTodoResponseObject, error) {
    uow := tenantrepo.NewTodoUnitOfWork(s.txManager, tenantSchema(req.Params.XTenantID))
    id, err := appcommand.NewCreateTodoHandler(uow).Handle(ctx, appcommand.CreateTodoCommand{
        Title: req.Body.Title,
    })
    if err != nil {
        return httpapi.CreateTodo500JSONResponse{N500JSONResponse: internalError(err)}, nil
    }
    return httpapi.CreateTodo201JSONResponse{Id: id}, nil
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

    // 4. Wiring warstw
    txManager := db.NewTxManager(pool)
    srv := handler.NewServer(txManager, pool)      // warstwa HTTP
    httpHandler := router.New(srv)                  // middleware + routing

    // 5. HTTP server
    httpServer := &http.Server{Addr: ":8080", Handler: httpHandler, ...}
    go httpServer.ListenAndServe()

    // 6. Czekaj na sygnał, graceful shutdown
    <-ctx.Done()
    httpServer.Shutdown(shutdownCtx)
}
```

### Dlaczego UoW jest tworzony per-request, nie raz przy starcie?

`TodoUnitOfWork` przechowuje `tenantSchema` — wartość specyficzną dla konkretnego requestu (z nagłówka `X-Tenant-ID`). Nie można go zbudować przy starcie aplikacji.

`TxManager` jest singletonem — tworzony raz, wstrzykiwany wszędzie.

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
                           — tworzy TodoUnitOfWork(txManager, "tenant_abc_123")
                           — tworzy CreateTodoHandler(uow)
                           — wywołuje handler.Handle(ctx, CreateTodoCommand{Title: "buy milk"})
       ↓
7. create_todo.go          — t := Todo{ID: uuid.New(), Title: "buy milk"}
                           — uow.Execute(ctx, fn)
       ↓
8. unit_of_work.go         — txManager.WithinTransaction(ctx, "tenant_abc_123", fn)
       ↓
9. tx.go                   — BEGIN
                           — SET LOCAL search_path = "tenant_abc_123"
                           — commandQ := commanddb.New(tx)
                           — fn(commandQ)
       ↓
10. todo_repository.go     — commandQ.CreateTodo(ctx, CreateTodoParams{ID: uuid, Title: "buy milk"})
       ↓
11. PostgreSQL             — INSERT INTO todos (id, title) VALUES (...)
                             (w schemacie tenant_abc_123)
       ↓
12. tx.go                  — COMMIT
       ↓
13. create_todo.go         — zwraca (id, nil)
       ↓
14. todo_handler.go        — zwraca CreateTodo201JSONResponse{Id: id}
       ↓
15. oapi-codegen           — serializuje do JSON, ustawia status 201
       ↓
16. Logging middleware     — loguje "request completed" {status: 201, duration: 3ms}
       ↓
17. HTTP response 201 {"id": "..."}
```

---

## 14. Kolejne kroki

### Krótkoterminowe
- [ ] Middleware autoryzacji — JWT, wyciąganie tenant ID z tokena zamiast z nagłówka
- [ ] `ListTodos` query handler — sqlc ma już `ListTodos` i `ListIncompleteTodos`
- [ ] Walidacja danych domenowych — konstruktory z logiką w encjach

### Średnioterminowe
- [ ] Testy unit dla HTTP handlerów — mapowanie błędów na kody HTTP z `httptest`
- [ ] Testy e2e — start serwera + prawdziwe HTTP requesty z testcontainers
- [ ] Obsługa błędów domenowych — własne typy błędów (np. `ErrNotFound`) zamiast `pgx.ErrNoRows` w handlerach

### Długoterminowe — CQRS i Event Sourcing
- [ ] Osobna baza dla read modeli — queries czytają z zoptymalizowanych projekcji
- [ ] Domain Events — command handler emituje zdarzenia (`TodoCreated`, `TodoCompleted`)
- [ ] Event Store — zdarzenia są źródłem prawdy, stan odtwarzany z eventów
- [ ] Outbox pattern — zdarzenia w tej samej transakcji co mutacja, potem asynchronicznie publikowane

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
