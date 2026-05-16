# CLAUDE.md — my-migrations

Wskazówki dla Claude Code przy pracy z tym projektem i przy tworzeniu kolejnych projektów o podobnej architekturze.

## Struktura katalogów

```
my-migrations/
├── cmd/
│   ├── api/main.go          # serwer HTTP — wiring, graceful shutdown
│   └── migrate/main.go      # CLI migracji: migrate | serve | retry-failed
├── internal/
│   ├── domain/              # modele, interfejsy repozytoriów, eventy, błędy
│   ├── application/
│   │   ├── command/         # handlery komend (zapis)
│   │   ├── query/           # handlery zapytań (odczyt)
│   │   └── port/            # interfejsy use-case'ów (wejście do aplikacji)
│   └── infrastructure/
│       ├── db/
│       │   ├── migrations/
│       │   │   ├── public/  # migracje schematu publicznego
│       │   │   └── tenant/  # migracje schematu tenanta
│       │   ├── queries/
│       │   │   ├── command/ # zapytania SQLC do zapisu (public/ i tenant/)
│       │   │   └── query/   # zapytania SQLC do odczytu (public/ i tenant/)
│       │   ├── sqlc/
│       │   │   ├── command/ # wygenerowany kod SQLC — package commanddb
│       │   │   └── query/   # wygenerowany kod SQLC — package querydb
│       │   ├── migrate.go   # logika migracji, walidacja, retry
│       │   ├── pool.go      # konfiguracja pgxpool
│       │   ├── tx.go        # TxManager — transakcje z routingiem schematu
│       │   └── txctx.go     # kontekst transakcji
│       ├── repository/      # implementacje interfejsów z domain/
│       ├── event/           # OutboxPublisher
│       ├── usecase/         # wiring: command/query handler + repo + publisher
│       └── worker/          # OutboxWorker — background processing
├── internal/http/
│   ├── api/                 # kod wygenerowany przez oapi-codegen
│   ├── handler/             # implementacja StrictServerInterface
│   ├── middleware/          # logging, request ID, recovery
│   └── router/              # Chi router z middleware chain
├── tests/integration/       # testy z testcontainers
├── charts/                  # Helm chart
│   ├── templates/
│   ├── values.yaml
│   ├── values.local.yaml    # środowisko lokalne (k3d)
│   └── values.dev.yaml      # środowisko dev
├── apis/                    # specyfikacje OpenAPI (YAML)
├── sqlc.yaml
├── Makefile
└── Dockerfile
```

## Warstwy architektury (DDD)

### Domain (`internal/domain/`)
- Encje z logiką biznesową, bez zależności od zewnętrznych pakietów
- Interfejsy repozytoriów (kontrakt danych, implementacja w infrastructure)
- Definicje eventów domenowych (`event.go`)
- Definicje błędów domenowych (`errors.go`)
- Nigdy nie importuje `infrastructure` ani `http`

### Application (`internal/application/`)
- **Command handlers** — operacje zapisu, przyjmują komendę, zwracają error
- **Query handlers** — operacje odczytu, przyjmują query, zwracają dane
- **Ports** — interfejsy use-case'ów (wejście HTTP → application)
- Zależność tylko od `domain/`, nie od `infrastructure/`

### Infrastructure (`internal/infrastructure/`)
- Implementacje repozytoriów z SQLC
- `usecase/` — wiring: łączy handlery komend/zapytań z repozytoriami i publisherem
- `db/` — pool, TxManager, migracje, wygenerowany kod SQLC
- `event/` — OutboxPublisher zapisujący eventy do tabeli `outbox_events`
- `worker/` — OutboxWorker pollujący `outbox_events` co N sekund

### HTTP (`internal/http/`)
- `handler/` — implementuje `StrictServerInterface` wygenerowany przez oapi-codegen
- Żadnej logiki biznesowej w handlerach — tylko: walidacja wejścia → wywołanie use-case'a → mapowanie odpowiedzi
- `router/` — Chi v5 z middleware chain: Recovery → Logging → OAPI Validation → RequestID → Handler

## Wielodostępność (multi-tenancy)

**Wzorzec: jeden schemat PostgreSQL na tenanta.**

- Tabele wspólne (`tenants`, `users`) w schemacie `public`
- Tabele prywatne (`todos`, `outbox_events`) w schemacie tenanta: `tenant_<uuid-z-myślnikami-zamienionymi-na-podkreślniki>`
- Każde żądanie wymaga nagłówka `X-Tenant-ID`
- `TxManager.WithinTransaction(ctx, tenantSchema, fn)` ustawia `SET LOCAL search_path = "tenant_schema"` wewnątrz transakcji

**Bezpieczeństwo:** nazwa schematu walidowana regexem `^[a-zA-Z0-9_]+$` przed interpolacją do SQL.

## TxManager — wzorzec transakcji

Cztery metody pokrywające wszystkie kombinacje:

```go
// Zapis w schemacie tenanta
txManager.WithinTransaction(ctx, tenantSchema, func(q *commanddb.Queries) error { ... })

// Zapis w schemacie publicznym
txManager.WithinPublicTransaction(ctx, func(q *commanddb.Queries) error { ... })

// Odczyt w schemacie tenanta (read-only tx)
txManager.WithinTransactionReadonly(ctx, tenantSchema, func(q *querydb.Queries) error { ... })

// Odczyt w schemacie publicznym (read-only tx)
txManager.WithinPublicTransactionReadonly(ctx, func(q *querydb.Queries) error { ... })
```

`commanddb` i `querydb` to dwa osobne pakiety SQLC — command do zapisu, query do odczytu.

## SQLC — dostęp do bazy danych

- Nie pisać ręcznie SQL w kodzie Go — tylko przez SQLC
- Zapytania w `internal/infrastructure/db/queries/command/` i `.../query/`
- Podzielone na `public/` (schemat publiczny) i `tenant/` (schemat tenanta)
- Po zmianie zapytań lub migracji: `make sqlc-generate`
- UUID mapowany do `github.com/google/uuid.UUID` (override w sqlc.yaml)
- Generuje: `models.go`, `querier.go`, `*_sql.go`, `db.go`

## Migracje — konwencje

### Nazewnictwo plików
Format: `000001_opis_tabeli.up.sql` / `000001_opis_tabeli.down.sql`

- Numery sekwencyjne (`000001`, `000002`, ...) — projekt solo
- Bez podkreślników w numerze wersji — golang-migrate parsuje wersję jako wszystko przed pierwszym `_`
- Osobne pliki dla `public/` i `tenant/`

### Nazewnictwo tabel
- Liczba mnoga: `tenants`, `users`, `todos`, `outbox_events`

### Walidacja (już zaimplementowana w `migrate.go`)
Przed uruchomieniem migracji sprawdzane są: luki w numeracji, duplikaty, puste pliki.

### Migracje osadzone w binarce
```go
//go:embed migrations/*
var MigrationsFS embed.FS
```
Nie kopiować plików migracji ręcznie — są osadzone w binarce.

### Przepływ migracji
1. `MigratePublic()` — tworzy tabele `tenants`, `users` w `public`
2. `MigrateAllTenants()` — dla każdego tenanta z `migration_status != 'success'`: tworzy schemat, uruchamia migracje, aktualizuje status
3. `MigrateSingleTenantWrapper()` — retry dla jednego tenanta z `migration_status = 'failed'`

## Outbox — wzorzec eventów

- Eventy domenowe zapisywane do `outbox_events` **wewnątrz tej samej transakcji** co dane biznesowe
- `OutboxWorker` polluje wszystkie schematy tenantów co N sekund (domyślnie 5s)
- Oznacza eventy jako opublikowane po przetworzeniu (`published_at`)
- Gwarantuje at-least-once delivery, odporne na crash

Gdy potrzebna integracja z Kafka/NATS — tylko zmienić implementację w `OutboxPublisher`, interfejs bez zmian.

## HTTP — kod generowany przez oapi-codegen

1. Specyfikacja OpenAPI w `apis/`
2. Generowanie: `make codegen` (lub `go generate ./...`)
3. Implementować `StrictServerInterface` w `internal/http/handler/`
4. Handler nie zna HTTP — przyjmuje i zwraca typy domenowe, oapi-codegen obsługuje marshalling

## cmd/api/main.go — wzorzec wiringu

Kolejność:
1. Logger (`slogctx` + `slog.NewJSONHandler`)
2. Signal context (`signal.NotifyContext` z SIGTERM, SIGINT)
3. Pool bazy danych (`db.NewPool`)
4. `TxManager`, `commanddb.New`, `querydb.New`, `OutboxPublisher`
5. `OutboxWorker` — `go worker.Run(ctx)`
6. Use-case'y z `usecase.NewXxx(...)`
7. `handler.NewServer(...)`, `router.New(...)`
8. `http.Server` z timeoutami
9. `<-ctx.Done()` → graceful shutdown z `http.Server.Shutdown()`

## cmd/migrate/main.go — wzorzec CLI

Tryby uruchamiane przez `os.Args[1]`:
- `migrate` — walidacja + MigratePublic + CreateTenant (dev) + MigrateAllTenants
- `serve` — prosty serwer HTTP (healthcheck)
- `retry-failed` — ponawia migracje dla tenantów z `migration_status = 'failed'`

Jeden binary obsługuje oba tryby — przełączane przez argument CLI.

## Helm / Kubernetes

- Migracja jako Helm hook `pre-install,pre-upgrade` — uruchamia `/app migrate` przed deployem aplikacji
- `wait-for-db` init container sprawdza dostępność bazy przed startem
- `migrate-retry-failed-job.yaml` — osobny Job do ręcznego retryu
- Środowiska: `values.local.yaml` (k3d), `values.dev.yaml` (dev)

## Testy integracyjne

- W `tests/integration/` z `testcontainers-go`
- Każdy test spinuje PostgreSQL w kontenerze Docker
- Uruchamia pełny stos migracji przed testami
- `make integration-test`
- Nie mockować bazy danych — testy muszą trafić w prawdziwą instancję PostgreSQL

## Konfiguracja środowiskowa

| Zmienna | Opis |
|---|---|
| `DATABASE_URL` | Pełny DSN (priorytet nad DB_*) |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` | Składowe DSN |
| `HTTP_ADDR` | Adres serwera (domyślnie `:8080`) |
| `APP_NAME` | Nazwa aplikacji (domyślnie `go-migrations`) |

## Kluczowe zależności

| Pakiet | Rola |
|---|---|
| `github.com/jackc/pgx/v5` | Driver PostgreSQL + connection pool |
| `github.com/golang-migrate/migrate/v4` | Migracje schematu |
| `github.com/sqlc-dev/sqlc` | Generowanie kodu z SQL (narzędzie, nie dep Go) |
| `github.com/oapi-codegen/oapi-codegen` | Generowanie kodu z OpenAPI |
| `github.com/go-chi/chi/v5` | Router HTTP |
| `github.com/google/uuid` | UUID v4 |
| `github.com/veqryn/slog-context` | Logger propagowany przez context |
| `github.com/testcontainers/testcontainers-go` | Kontenery w testach |

## Czego nie robić

- Nie pisać SQL bezpośrednio w kodzie Go — używać SQLC
- Nie interpolować nazwy schematu bez walidacji regexem
- Nie umieszczać logiki biznesowej w handlerach HTTP
- Nie importować `infrastructure` z `domain`
- Nie mockować bazy danych w testach integracyjnych
- Nie zmieniać nazw zaaplikowanych plików migracji (zmiana numeru wersji = błąd przy down migration)
