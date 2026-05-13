Podsumowanie sesji

Co zrobiliśmy

Kontynuacja refaktoru ReadRepository (kroki 1–4 z sesji 2026-05-12).
Ukończono cały zaległy dług techniczny i dodano nowe wzorce DDD.

---

ReadRepository refaktor (domknięcie z poprzedniej sesji)

  Krok 1 — internal/domain/todo/read_repository.go — interface ReadRepository {GetByID, List}
  Krok 2 — GetTodoHandler i ListTodosHandler używają todo.ReadRepository zamiast todo.Repository
  Krok 3 — internal/infrastructure/repository/tenant/todo_read_repository.go — typ TodoReadRepository
    chowa WithinTransactionReadonly w każdej metodzie
  Krok 4 — GetTodo i ListTodos w todo_handler.go uproszczone do 2 linijek (brak logiki transakcji)

ErrConflict (409) dla CreateTenant i CreateUser

  Mapowanie pgconn.PgError kod "23505" → domain.ErrConflict w repozytorium
  HTTP handler: errors.Is(err, domain.ErrConflict) → 409JSONResponse
  Helper conflictError() w handler/errors.go

CreatedAt w TodoResult i GetTodoResult

  Dodano pole CreatedAt time.Time do obu DTO
  HTTP odpowiedź przestała zwracać null dla created_at

Testy integracyjne

  tests/integration/todo_query_handler_test.go — TestListTodosHandler_Handle
  tests/integration/conflict_test.go — TestCreateTenantHandler_ConflictOnDuplicateSchema,
    TestCreateUserHandler_ConflictOnDuplicateEmail
  tests/integration/todo_http_test.go — pełne pokrycie HTTP: CreateTodo, GetTodo (200/404),
    ListTodos, CompleteTodo, DeleteTodo — przez httptest.NewServer + router.New
  tests/integration/tenant_user_http_test.go — CreateTenant, CreateTenant_Conflict,
    GetTenant (200/404), CreateUser, CreateUser_Conflict, GetUser (200/404)

Unit testy

  internal/domain/user/email_test.go — testy value objectu Email
  internal/application/query/get_todo_test.go — TestGetTodoHandler_Handle_ReturnsNotFound,
    TestGetTodoHandler_Handle_ReturnsMappedTodo
  FakeTodoRepo.GetByID naprawiony — szuka po ID w Created, zwraca domain.ErrNotFound gdy brak

---

Nowe wzorce DDD

Aggregate z inwariantami (commit: My go topics - DDD aggregate with invariants)

  todo.NewTodo(id, title) (*Todo, error) — waliduje tytuł (nie może być pusty/whitespace)
    → domain.ErrInvalidTitle
  (*Todo).Complete() error — zwraca domain.ErrAlreadyCompleted jeśli już ukończone
  todo.Repository zmienione: usunięto Complete(id), dodano GetByID + Update
  CompleteTodoHandler: load → t.Complete() → repo.Update(*t)
  TodoRepository.Update używa istniejącego CompleteTodo SQL (jedyne mutowalne pole to Completed)
  Nowe błędy domenowe: ErrAlreadyCompleted, ErrInvalidTitle
  HTTP: CreateTodo z pustym tytułem → 400 (CreateTodo400JSONResponse)
  Brak 409 dla CompleteTodo w OpenAPI spec — do poprawki w przyszłości

Email value object (commit: My go topics - DDD Email value object)

  internal/domain/user/email.go — typ Email{value string} (pole prywatne)
  NewEmail(s string) (Email, error) — normalizuje (lowercase + trim), waliduje format
    (jeden @, domena z kropką)
  Email.String() string — jedyny accessor
  Propagacja: User.Email to Email, nie string; repo używa .String() do SQL i NewEmail() przy odczycie
  Nowy błąd: ErrInvalidEmail
  HTTP: CreateUser z nieprawidłowym emailem → 400
  Testy: email_test.go (valid/invalid/normalizacja), create_user_test.go (NormalizesEmail, InvalidEmail)

Domain events (commit: My go topics - DDD domain events (TodoCreated))

  internal/domain/event.go — interfejsy DomainEvent i EventPublisher
  internal/domain/todo/events.go — TodoCreated{TodoID, Title, OccurredAt}
  Agregat nagrywa eventy wewnętrznie: NewTodo wywołuje t.record(TodoCreated{...})
  PullEvents() []DomainEvent — zwraca i czyści listę eventów
  CreateTodoHandler: po udanej transakcji → t.PullEvents() → publisher.Publish()
  internal/infrastructure/event/log_publisher.go — LogPublisher (slog) na produkcję
  testhelpers.FakeEventPublisher — zbiera eventy do asercji w testach
  Komentarz w create_todo.go wyjaśnia: w produkcji OutboxPublisher powinien pisać do DB
    w tej samej transakcji dla gwarantowanej dostawy
  Server.eventPublisher wstrzykiwany przez NewServer(txManager, pool, publisher)

---

Stan na jutro

Wszystko zacommitowane (10 commitów ahead of origin/main).

---

Kolejne kroki DDD

1. Input Ports — interfejsy warstwy aplikacyjnej (zamknięcie hexagonal architecture)

   Aktualny problem: HTTP handler instancjonuje konkretny struct command.NewCreateTodoHandler(...) per-request.
   To wyciek — HTTP zna szczegóły konstrukcji application layer.

   Rozwiązanie: interfejsy w internal/application/port/:
     type CreateTodoUseCase interface { Handle(ctx, cmd) (uuid.UUID, error) }
     type CompleteTodoUseCase interface { Handle(ctx, cmd) error }
     itd.

   Server dostaje zbudowane use case'y przez konstruktor NewServer(...) — nie tworzy ich per-request.
   Korzyść: HTTP handler jest unit-testowalny — przekazujesz mock zamiast handlera z bazą.

   Priorytet: WYSOKI — to jedyna brakująca część hexagonal architecture.

2. ErrAlreadyCompleted → 409 w OpenAPI spec

   Mały dług: CompleteTodo nie ma odpowiedzi 409 w mymigrations.yaml.
   ErrAlreadyCompleted wraca jako 500 zamiast 409.
   Prosta zmiana: dodać 409 do spec → go generate → obsłużyć w handlerze.

   Priorytet: ŚREDNI — poprawność API.

3. TodoCompleted / TodoDeleted domain events

   Agregat powinien nagrywać eventy w metodach Complete() i Delete() (analogicznie do TodoCreated).
   Wymaga: nowych typów eventów w events.go, wywołania t.record(...) w metodach agregatu,
   PullEvents() w handlerach kompletowania i usuwania.

   Priorytet: ŚREDNI — spójność wzorca.

4. OutboxPublisher

   Gwarantowana dostawa eventów: publisher zapisuje do tabeli outbox w tej samej transakcji co agregat.
   Wymaga: nowej migracji (tabela outbox), implementacji OutboxPublisher w infrastructure/event/,
   osobnego procesu (goroutine lub cron) czytającego outbox i publikującego dalej.

   Priorytet: NISKI na tym etapie nauki — złożona infrastruktura, warto zrozumieć wzorzec teorią.
