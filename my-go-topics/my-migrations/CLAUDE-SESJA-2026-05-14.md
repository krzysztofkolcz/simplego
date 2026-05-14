Podsumowanie sesji

Co zrobiliśmy

Domknięcie hexagonal architecture przez Input Ports, poprawka 409 dla ErrAlreadyCompleted,
domain events dla TodoCompleted i TodoDeleted.

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

Stan na jutro

Wszystko zacommitowane. 5 nowych commitów (dziś) ahead of origin/main.
DDD-CLAUDE.md zaktualizowany o Input Ports (sekcja 18), sekcję 16 (Domain Events),
sekcję 14 (Aggregate) i sekcję 19 (Kolejne kroki).

---

Kolejne kroki

1. OutboxPublisher — NISKI priorytet

   Gwarantowana dostawa eventów: publisher zapisuje do tabeli outbox
   w tej samej transakcji co agregat.
   Wymaga: nowej migracji (tabela outbox), implementacji OutboxPublisher
   w infrastructure/event/, osobnego procesu czytającego outbox i publikującego dalej.

2. Unit testy HTTP handlerów — ŚREDNI priorytet

   Teraz możliwe bez bazy: Server dostaje fake port interface.
   Testować: mapowanie błędów domenowych na kody HTTP (400, 404, 409, 500).
   Plik: internal/http/handler/todo_handler_test.go

3. Middleware autoryzacji — NISKI priorytet na tym etapie

   JWT, wyciąganie tenant ID z tokena zamiast z nagłówka X-Tenant-ID.
