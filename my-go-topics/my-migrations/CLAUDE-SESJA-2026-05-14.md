Podsumowanie sesji

Co zrobiliśmy

Domknięcie hexagonal architecture przez Input Ports, poprawka 409 dla ErrAlreadyCompleted,
domain events dla TodoCompleted i TodoDeleted, unit testy HTTP handlerów z fake portami,
OutboxPublisher z gwarantowaną dostawą eventów.

---

Input Ports — hexagonal architecture (commit: My go topics - DDD Input Ports)

  Problem: HTTP handler tworzył per-request konkretne typy infrastruktury
    (tenantrepo.NewTodoUnitOfWork, appcommand.NewCreateTodoHandler)
    i importował infrastructure/repository/tenant oraz infrastructure/db.
    To wyciek — warstwa HTTP znała szczegóły konstrukcji warstwy aplikacyjnej.

  Rozwiązanie: interfejsy use case w internal/application/port/

    type CreateTodoPort interface { Handle(ctx, cmd) (uuid.UUID, error) }
    type CompleteTodoPort interface { Handle(ctx, cmd) error }
    type DeleteTodoPort interface { Handle(ctx, cmd) error }
    type GetTodoPort interface { Handle(ctx, q) (*GetTodoResult, error) }
    type ListTodosPort interface { Handle(ctx, q) ([]TodoResult, error) }
    type CreateTenantPort, GetTenantPort, CreateUserPort, GetUserPort — analogicznie

  Use case adaptery w internal/infrastructure/usecase/
    Trzy pliki: todo_use_case.go, tenant_use_case.go, user_use_case.go
    Każdy adapter implementuje port interface, tworzy UoW/repo na podstawie cmd.TenantSchema
    i deleguje do istniejącego application handler.

    func (u *createTodoUseCase) Handle(ctx, cmd) (uuid.UUID, error) {
        uow := tenantrepo.NewTodoUnitOfWork(u.txManager, cmd.TenantSchema)
        return appcommand.NewCreateTodoHandler(uow, u.eventPublisher).Handle(ctx, cmd)
    }

  TenantSchema string dodane do tenant-scoped command/query:
    CreateTodoCommand, CompleteTodoCommand, DeleteTodoCommand,
    GetTodoQuery, ListTodosQuery

  Server struct — zero importów infrastruktury w http/handler/:
    type Server struct {
        createTodo   port.CreateTodoPort
        completeTodo port.CompleteTodoPort
        // ... 9 pól łącznie
    }

  NewServer(txManager, pool, publisher) → NewServer(9 port interfaces)
  main.go i newTestServer() w testach — wiring use case adapterów

  Korzyść: HTTP handler można unit-testować z fake portami, bez bazy danych.

---

ErrAlreadyCompleted → 409 (commit: My go topics - DDD ErrAlreadyCompleted → 409)

  apis/mymigrations.yaml — dodano "409": $ref: "#/components/responses/409" do completeTodo
  oapi-codegen (ręcznie z roota: oapi-codegen -config apis/config.yaml apis/mymigrations.yaml)
    → wygenerowany CompleteTodo409JSONResponse
  todo_handler.go — errors.Is(err, domain.ErrAlreadyCompleted) → CompleteTodo409JSONResponse
    Wiadomość z err.Error() ("already completed"), nie z conflictError() ("Resource already exists")

---

Domain Events TodoCompleted i TodoDeleted (commit: My go topics - DDD domain events)

  events.go — dwa nowe typy:
    type TodoCompleted struct { TodoID uuid.UUID; OccurredAt time.Time }
    type TodoDeleted   struct { TodoID uuid.UUID; OccurredAt time.Time }

  entity.go:
    Complete() — po zmianie stanu: t.record(TodoCompleted{...})
    Delete() — nowa metoda void (brak inwariantów), tylko: t.record(TodoDeleted{...})

  complete_todo.go — dodano publisher domain.EventPublisher do konstruktora
    var t *todo.Todo zadeklarowane poza closure → t.PullEvents() po uow.Execute

  delete_todo.go — handler teraz ładuje agregat przed usunięciem:
    var t *todo.Todo
    uow.Execute: t, err = repo.GetByID → t.Delete() → repo.Delete(cmd.ID)
    po commicie: publisher.Publish(ctx, t.PullEvents())

    Dlaczego GetByID przed Delete: agregat jest źródłem eventów.
    Bez załadowania nie ma obiektu t do PullEvents.
    Dodatkowa korzyść: ErrNotFound na poziomie domeny, nie bazy.

  infrastructure/usecase/ — NewCompleteTodoUseCase i NewDeleteTodoUseCase
    dostają teraz publisher jako drugi argument

  Testy:
    complete_todo_test.go — pub *FakeEventPublisher wstrzyknięty; asercja pub.Published[0].EventName()
    delete_todo_test.go — TestDeleteTodoHandler_Handle_DeletesTodo wymaga wpisu w repo.Created
      (handler robi GetByID przed Delete); asercja na "todo.deleted"

---

Unit testy HTTP handlerów z fake portami (commit: My go topics - DDD unit testy HTTP handlerów)

  25 testów, 19ms, zero połączeń z bazą danych.

  Struktura — 4 pliki w internal/http/handler/ (package handler):
    fakes_test.go          — 9 fake struct implementujących port interfaces
    todo_handler_test.go   — 14 testów (CreateTodo ×3, GetTodo ×3, CompleteTodo ×4,
                             DeleteTodo ×2, ListTodos ×2)
    tenant_handler_test.go — 5 testów (CreateTenant ×3, GetTenant ×2)
    user_handler_test.go   — 5 testów (CreateUser ×3, GetUser ×2)

  Wzorzec fake portu — prosta struct z konfigurowalnymi polami:
    type fakeCreateTodo struct { id uuid.UUID; err error }
    func (f *fakeCreateTodo) Handle(_ context.Context, _ appcommand.CreateTodoCommand) (uuid.UUID, error) {
        return f.id, f.err
    }

  Wzorzec testu — fake wstrzyknięty bezpośrednio w pole Server:
    ts := todoSrv(&Server{completeTodo: &fakeCompleteTodo{err: domain.ErrAlreadyCompleted}})
    // → assert 409

  &Server{completeTodo: fake} działa, bo testy są w package handler
  (dostęp do nieeksportowanych pól). Pozostałe pola nil — nie są wywoływane.

  Pokrycie: złota ścieżka (200/201/204) + każde mapowanie błędu domenowego:
    ErrInvalidTitle → 400, ErrInvalidEmail → 400
    ErrNotFound → 404
    ErrAlreadyCompleted → 409, ErrConflict → 409
    errors.New("db down") → 500
  Dla 200/201: asercja na pola JSON (id, title, schema_name, email).

---

OutboxPublisher (commit: My go topics - DDD OutboxPublisher)

  Problem: publisher.Publish wywoływany PO uow.Execute — crash między commitem
  a publishem powoduje utratę eventu. LogPublisher tym nie przejmuje, ale OutboxPublisher musi
  pisać do bazy w TEJ SAMEJ transakcji co agregat.

  Kluczowa zmiana architekturalna — todo.UnitOfWork.Execute fn dostaje teraz ctx:

    // przed:
    Execute(ctx context.Context, fn func(repo Repository) error) error

    // po:
    Execute(ctx context.Context, fn func(ctx context.Context, repo Repository) error) error

  Dlaczego: fn musi dostać txCtx (z *commanddb.Queries osadzonym w context),
  żeby OutboxPublisher mógł wyciągnąć zapytania i wstawić do outbox_events.

  Przepływ:
    1. TodoUnitOfWork.Execute wywołuje txManager.WithinTransaction
    2. Wewnątrz otrzymuje commandQ dla aktywnej transakcji
    3. txCtx = db.ContextWithTxQueries(ctx, commandQ)
    4. fn(txCtx, repo) — fn jest closurem handlera
    5. Handler: repo.Update(txCtx, *t) + publisher.Publish(txCtx, t.PullEvents())
    6. OutboxPublisher.Publish: q, ok := db.TxQueriesFromCtx(ctx) → INSERT INTO outbox_events
    7. Transakcja commituje — agregat i event w outbox atomicznie

  Nowe pliki:
    internal/infrastructure/db/txctx.go
      db.ContextWithTxQueries(ctx, q) → context.Context
      db.TxQueriesFromCtx(ctx) → (*commanddb.Queries, bool)

    internal/infrastructure/event/outbox_publisher.go
      type OutboxPublisher struct{}
      Publish: jeśli brak txCtx → zwróć nil (np. w testach jednostkowych)
      W transakcji: json.Marshal(event) → q.InsertOutboxEvent

  Zmienione pliki:
    internal/domain/todo/unit_of_work.go — fn dostaje ctx
    internal/infrastructure/repository/tenant/unit_of_work.go — przekazuje txCtx do fn
    internal/testhelpers/fake_repo.go — FakeTodoUoW.Execute dostosowane
    internal/application/command/create_todo.go — fn dostaje ctx, publish wewnątrz
    internal/application/command/complete_todo.go — fn dostaje ctx, publish wewnątrz
    internal/application/command/delete_todo.go — fn dostaje ctx, publish wewnątrz

  Ważna właściwość: publisher.Publish(txCtx, events) wywoływane WEWNĄTRZ fn
  (czyli wewnątrz transakcji). Jeśli brak *commanddb.Queries w ctx (np. fake UoW
  w testach jednostkowych), OutboxPublisher zwraca nil — żadnego efektu.
  Eventy trafiają do outbox tylko gdy jest prawdziwa transakcja.

---

outbox_events w schemacie tenanta + testy integracyjne
(commit: My go topics - outbox_events w schemacie tenanta + testy integracyjne)

  Decyzja: outbox_events per-tenant (nie w public)

  Pierwsza wersja OutboxPublisher pisała do public.outbox_events, INSERT wywoływany
  wewnątrz transakcji tenanta (SET LOCAL search_path = "tenant_xxx"). Tabela public.outbox_events
  była niedostępna — search_path zasłaniał ją schemątem tenanta.

  Rozwiązanie: outbox_events w każdym schemacie tenanta.
    - SET LOCAL search_path = "tenant_xxx" pokrywa outbox_events automatycznie
    - brak potrzeby kwalifikowania zapytań przez public.
    - Worker iteruje per-tenant: ListTenantSchemas → processTenant

  Przeniesione pliki:
    migrations/public/002_outbox.* → migrations/tenant/002_outbox.*
    queries/command/public/outbox.sql → queries/command/tenant/outbox.sql
    (zapytania bez prefiksu public. — search_path załatwia schemę)

  OutboxWorker — przepisany na per-tenant:
    type OutboxWorker struct { txManager *db.TxManager; publicQ *commanddb.Queries; interval }
    ProcessOnce: publicQ.ListTenantSchemas(ctx) → for schema → processTenant(ctx, schema)
    processTenant: txManager.WithinTransaction(ctx, schema, fn)
      fn: SelectUnpublishedOutboxEvents → log → MarkOutboxEventPublished

    Wcześniej worker operował na jednej tabeli public.outbox_events.
    Teraz musi iterować wszystkich tenantów i procesować każdego osobno.

  Wiring w main.go:
    eventPublisher := infraevent.NewOutboxPublisher()
    outboxWorker := worker.NewOutboxWorker(txManager, commandQ, 5*time.Second)
    go outboxWorker.Run(ctx)

  Testy integracyjne (tests/integration/outbox_test.go) — 5 testów:
    TestOutboxPublisher_CreateTodo_WritesEventToOutbox
      — truncate outbox, CreateTodo przez OutboxPublisher, assert 1 event "todo.created"
        z polem TodoID i Title w payload
    TestOutboxPublisher_CompleteTodo_WritesEventToOutbox
      — create (LogPublisher), truncate, CompleteTodo (OutboxPublisher), assert "todo.completed"
    TestOutboxPublisher_DeleteTodo_WritesEventToOutbox
      — create, truncate, DeleteTodo (OutboxPublisher), assert "todo.deleted"
    TestOutboxPublisher_LogPublisher_DoesNotWriteToOutbox
      — truncate, create (LogPublisher), assert empty outbox
    TestOutboxWorker_ProcessOnce_MarksEventsPublished
      — truncate, create event, assert 1 unpublished, ProcessOnce, assert 0 unpublished,
        assert published_at != nil

  Pomocniki testowe operują na schemacie tenanta:
    truncateOutbox: TRUNCATE "tenant_xxx".outbox_events RESTART IDENTITY
    unpublishedEvents: WithinTransaction(TenantSchema) → SelectUnpublishedOutboxEvents
    allOutboxEvents: SELECT ... FROM "tenant_xxx".outbox_events

  Naprawione pre-existing bugs:
    MigrateAllTenants — pomija schematy z public.tenants które nie istnieją w PostgreSQL
      (fix: TestCreateTenantHandler_ConflictOnDuplicateSchema wstawia "duplicate_schema"
       bez tworzenia schematu; MigrateAllTenants próbował go migrować i failował)
    TestTenantTableWorks — używa SET LOCAL search_path w transakcji zamiast SET search_path
      (fix: SET search_path bez LOCAL zatruwało pulę połączeń; kolejne testy dostawały
       połączenie z search_path = "tenant_xxx" i nie mogły znaleźć public.tenants)
    todo_repository_test.go, todo_command_handler_test.go — pgx.ErrNoRows → domain.ErrNotFound

  Stan testów: 33/33 pass (integration) + wszystkie unit testy

---

Stan na koniec sesji

Wszystko zacommitowane. 13 commitów ahead of origin/main.

---

Kolejne kroki

1. Middleware autoryzacji — NISKI priorytet na tym etapie

   JWT, wyciąganie tenant ID z tokena zamiast z nagłówka X-Tenant-ID.
