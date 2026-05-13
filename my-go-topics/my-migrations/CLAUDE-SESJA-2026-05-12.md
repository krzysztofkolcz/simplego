Podsumowanie sesji

Co zrobiliśmy

Kontynuacja z sesji 2026-05-11. Zaimplementowaliśmy ListTodos end-to-end, zrefaktorowaliśmy
testy i obsługę błędów, stworzyli dokumentację przepływu.

testhelpers (internal/testhelpers/fake_repo.go) — przeniesiono wszystkie fake'i
(FakeTodoRepo, FakeTodoUoW, FakeUserRepo, FakeUserUoW, FakeTenantRepo, FakeTenantUoW)
z fakes_test.go w pakiecie command do wspólnego pakietu testhelpers z publicznymi nazwami.
Stary fakes_test.go usunięty. Testy w command i query korzystają z jednego źródła.

Unit testy dla ListTodosHandler (internal/application/query/list_todos_test.go) —
dwa testy: pusty wynik i mapowanie encji.

Domain error types (internal/domain/errors.go) — dodano ErrNotFound.
TodoRepository mapuje pgx.ErrNoRows → domain.ErrNotFound w GetByID, Complete, Delete.
HTTP handlery używają domain.ErrNotFound zamiast pgx.ErrNoRows — usunięto import pgx z handlera.

FLOW.md — plik dokumentujący pełny przepływ od HTTP handlera przez TxManager, UnitOfWork,
Repository do wygenerowanego kodu sqlc i PostgreSQL. Zawiera dwa flow (zapis i odczyt),
wyjaśnienie podziału command/query, tabelę plików.

---
Stan na jutro

Niezcommitowane zmiany — wszystkie pliki z sesji gotowe do commita.

Refaktoring w toku — ReadRepository (krok 1 zrobiony częściowo):

  Krok 1 — dodać ReadRepository interface do internal/domain/todo/repository.go:
    type ReadRepository interface {
        GetByID(ctx context.Context, id uuid.UUID) (*Todo, error)
        List(ctx context.Context) ([]Todo, error)
    }

  Krok 2 — zmienić GetTodoHandler i ListTodosHandler żeby używały todo.ReadRepository
    zamiast todo.Repository (internal/application/query/get_todo.go i list_todos.go)

  Krok 3 — nowy plik internal/infrastructure/repository/tenant/todo_read_repository.go
    Typ TodoReadRepository implementuje todo.ReadRepository, chowa WithinTransactionReadonly
    w środku każdej metody. HTTP handler nie będzie już wywoływał TxManager bezpośrednio.

  Krok 4 — uprościć GetTodo i ListTodos w todo_handler.go:
    repo := tenantrepo.NewTodoReadRepository(s.txManager, tenantSchema(req.Params.XTenantID))
    result, err := appquery.NewGetTodoHandler(repo).Handle(ctx, appquery.GetTodoQuery{ID: req.Id})

Znany dług techniczny:
  - ErrNotFound brakuje w tenant i user repository (analogiczny refaktor)
  - ErrConflict dla CreateTenant/CreateUser przy duplikacie (409 Conflict)
  - Integracyjny test dla ListTodos endpoint
  - TodoResult w list_todos.go nie ma pola CreatedAt (przekazywane nil do HTTP)
