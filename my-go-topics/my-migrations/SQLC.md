# Jak połączyć migracje oraz multitenancy z sqlc.
https://chatgpt.com/g/g-p-6974db69d3dc819196dfb45bdb7bd10a-saas/c/69fcbef5-5474-8395-9879-0f7c2ab16d3d

# generowanie
sqlc generate
# struktura
```
project/
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   │
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   │   │
│   │   ├── db/
│   │   │   │
│   │   │   ├── migrations/
│   │   │   │   │
│   │   │   │   ├── public/
│   │   │   │   │   ├── 000001_create_tenants.up.sql
│   │   │   │   │   └── 000001_create_tenants.down.sql
│   │   │   │   │
│   │   │   │   └── tenant/
│   │   │   │       ├── 000001_create_customers.up.sql
│   │   │   │       ├── 000001_create_customers.down.sql
│   │   │   │       ├── 000002_create_orders.up.sql
│   │   │   │       └── ...
│   │   │   │
│   │   │   ├── queries/
│   │   │   │   │
│   │   │   │   ├── customer.sql
│   │   │   │   ├── order.sql
│   │   │   │   └── ...
│   │   │   │
│   │   │   ├── sqlc/ # Generowane przez sqlc
│   │   │   │   │
│   │   │   │   ├── db.go
│   │   │   │   ├── models.go
│   │   │   │   ├── querier.go
│   │   │   │   ├── customer.sql.go
│   │   │   │   ├── order.sql.go
│   │   │   │   └── batch.go
│   │   │   │
│   │   │   ├── tx.go
│   │   │   ├── tenant.go # OK - WithTenant
│   │   │   ├── migrate.go
│   │   │   └── pool.go
│   │   ├── repository/
│   │   │   ├── customer_repository.go
│   │   │   └── order_repository.go
│   │   │
│   │   └── messaging/
│   │
│   ├── interfaces/
│   │   └── http/
│   │       ├── middleware/
│   └── shared/
│
├── sqlc.yaml
├── Makefile
├── go.mod
└── docker-compose.yml
```

# Idealna struktura
```
internal/
└── infrastructure/
    └── db/
        │
        ├── migrations/
        │   ├── public/
        │   └── tenant/
        │
        ├── queries/
        │   ├── customer.sql
        │   ├── invoice.sql
        │   └── ...
        │
        ├── sqlc/
        │   ├── models.go
        │   ├── db.go
        │   ├── querier.go
        │   └── *.sql.go
        │
        ├── pool.go
        ├── tx.go
        ├── migrate.go
        ├── tenant.go
        └── testing.go
sqlc.yaml
```
migracje i db w roocie - dobre dla małych projektów.


## pool.go

Tworzenie:

pgxpool.Pool

oraz:

retry
tracing
metrics
config

Dlaczego pool jest ważny przy multitenancy
Bo:

SET search_path

jest ustawiany na CONNECTION.

A pool reużywa connectiony.

Dlatego:

✅ tenant context MUSI być ustawiany per transaction
❌ nigdy globalnie


### Co robi PrepareConn?

Uruchamia się:

tuż przed przekazaniem połączenia z poola

czyli dobre miejsce na:

validation
reset state
metrics
sanity checks


## tx.go

Transaction manager.

Tu będzie:

WithinTransaction
SET LOCAL search_path
tenant-aware tx

```
HTTP request
    ↓
middleware
    ↓
tenant resolved
    ↓
command handler
    ↓
TxManager.WithinTransaction(...)
    ↓
SET LOCAL search_path
    ↓
repositories/sqlc
```

### Najważniejsza zasada

NIGDY:

❌ nie ustawiaj tenant schema globalnie
❌ nie ustawiaj tenant schema na pool
❌ nie ustawiaj tenant schema poza transaction scope

ZAWSZE:

✅ SET LOCAL search_path

wewnątrz transakcji.

Dlaczego SET LOCAL?

To kluczowe.

SET LOCAL

działa tylko:

w obrębie transakcji

Po:

COMMIT

lub:

ROLLBACK

Postgres automatycznie resetuje setting.

To production-grade sposób.

### pgx.Identifier
Dlaczego pgx.Identifier?
NIE rób:
fmt.Sprintf("SET LOCAL search_path = %s", tenant)
Bo:
❌ SQL injection risk
Lepiej:
pgx.Identifier{tenant}.Sanitize()

### Nested transaction
Czy robić nested transactions?

Na początku:

❌ NIE

To komplikuje:

savepoints
rollback semantics
orchestration

Jedna transaction boundary per command handler jest idealna.

## tenant.go
Wykorzystanie:
```
tenant := db.Tenant{
	ID: "tenant_123",
	Schema: "tenant_123",
}

ctx = db.ContextWithTenant(ctx, tenant)
```
```
tenant, err := db.TenantFromContext(ctx)
```

## config.go
Konfiguracja, wykorzystanie:
```
cfg := db.DefaultConfig()

cfg.DatabaseURL = os.Getenv("DATABASE_URL")

if err := cfg.Validate(); err != nil {
	log.Fatal(err)
}
```
Dlaczego osobny config package?

Bo później możesz dodać:

✅ envconfig
✅ viper
✅ koanf
✅ secret manager
✅ dynamic reload

bez ruszania pool.go.


## Sqlc config
```
my-migrations/
├── sqlc.yaml
```
## queries/
./internal/infratsructure/db/queries/
Ręcznie pisane SQL dla sqlc.
Proponowana struktura:
```
internal/
└── infrastructure/
    └── db/
        |
        .
        .
        .
        ├── queries/
        │   ├── command/
        │   │   ├── public/
        │   │   │   ├── tenants.sql
        │   │   │   └── users.sql
        │   │   │
        │   │   └── tenants/
        │   │       └── todos.sql
        │   │
        │   └── query/
        │       ├── public/
        │       │   ├── tenants.sql
        │       │   └── users.sql
        │       │
        │       └── tenants/
        │           └── todos.sql
```

## sqlc/
Tylko generated files.
Nigdy ręcznie nie edytujesz.
```
sqlc/
├── db.go
├── models.go
├── querier.go
└── *.sql.go
```
lub rozdzielenie command / query
```
sqlc/
├── command/
└── query/
```
## tenant.go

Tenant context helpers:

type Tenant struct {
    ID     string
    Schema string
}

## migrate.go

Odpalanie:

global migrations
tenant migrations


## public.
Chat sugeruje, żeby sqlc zapytania do public zawsze robić z przedrostkiem public.


## pgxpool vs sql.DB
✅ używaj pgxpool gdy:
piszesz backend (Twoje API)
robisz:
batch
COPY
performance tuning
✅ używaj sql.DB gdy:
używasz:
golang-migrate
ORM
legacy libs

🔥 Best practice (Twoja sytuacja)

👉 używaj OBU

// runtime app
pgxpool

// migrations
sql.DB
💡 Pro tip

Jeśli chcesz połączyć światy:

import "github.com/jackc/pgx/v5/stdlib"

db := stdlib.OpenDB(*pgxConfig)
### pgxpool
```
Twoja aplikacja
↓
pgxpool
↓
pgx
↓
PostgreSQL
```
conn, _ := pool.Acquire(ctx)
defer conn.Release()

conn.Exec(ctx, "SET search_path...")
conn.Exec(ctx, "SELECT ...")
masz pewność, że to TEN SAM connection

### sql.DB - database/sql
```
Twoja aplikacja
↓
database/sql
↓
pgx (stdlib wrapper)
↓
PostgreSQL
```
database/sql:
db.Exec("SET search_path TO ...")

👉 działa na jednym connection

ALE:

db.Query(...)

👉 może użyć innego connection 😐

## Docelowa architektura migracji
🔥 Docelowa architektura (SaaS)
zamiast:
deploy → migrate all tenants
robisz:
deploy → migrate public
       → enqueue tenants
       → background worker migrates tenants

opcja	        kiedy
sekwencyjnie	mała skala
batch	        średnia skala
worker pool	    duża skala
async system	SaaS production

### Moja rekomendacja dla Ciebie
Na teraz:
👉 worker pool (5–10 workerów)
Na przyszłość:
👉 background migration system (outbox + worker)
### Na przyszłość
🔁 system migracji tenantów jako job queue
📦 retry + backoff per tenant
📊 monitoring migracji (metrics)

To jest dokładnie poziom „SaaS jak Stripe / Shopify” 👍

### Status migracji
pending → running → success / failed

### Retry job
migrate-retry-failed-job.yaml

main.go
retry-failed

### Monitoring sql
-- ile sukcesów
SELECT migration_status, COUNT(*)
FROM tenants
GROUP BY migration_status;

-- które padły
SELECT * FROM tenants WHERE migration_status = 'failed';

### Procedura po błedzie migracji
1. sprawdź:
   SELECT * FROM tenants WHERE migration_status = 'failed';

2. popraw migrację / dane

3. odpal:
   kubectl create job retry-tenants ...

lub:

   /app retry-failed

### Dirty database
⚠️ 7. Edge cases (ważne)
❗ Dirty DB

log:

Dirty database version X

👉 wtedy:

ręcznie:
DELETE FROM tenant_xxx.schema_migrations WHERE dirty = true;





# TODO
Observability

OpenTelemetry:

otelpgx
Slow query logging
Query tracing
Pool metrics

Bardzo ważne.

Np:

stats := pool.Stat()

i Prometheus metrics:

acquired conns
idle conns
wait duration
acquire count
9. Czy robić singleton?

NIE.

Twórz jawnie:

pool := db.NewPool(...)

i dependency injection.

10. Czy robić wrapper na pool?

Tak, ale mały.

Np:

type Database struct {
	Pool *pgxpool.Pool
}

Później możesz tam dodać:

metrics
tx manager
repositories factory
health checks

Mogę Ci pokazać:

🔁 system retry (cron / job queue)
📊 metryki migracji (ile success/fail)
🧱 izolację migracji (żeby jeden tenant nie wpływał na innych)

To jest dokładnie poziom „stabilny SaaS backend” 👍

Możemy dorobić:

🔁 kolejkę migracji (Kafka / Redis)
⚡ worker pool
📈 metryki Prometheus
🔒 migration lock

To już poziom „Shopify-grade SaaS infra” 👍

🚀 Jeśli chcesz iść dalej

Mogę Ci pokazać:

jak napisać własny driver (10x lepsze zrozumienie)
jak działa locking w migracjach (bardzo ważne w produkcji)
albo jak zrobić migracje bez golang-migrate

To już poziom „rozumiem system od środka” 👍

# DDD 
https://chatgpt.com/g/g-p-6974db69d3dc819196dfb45bdb7bd10a-saas/c/69fcbef5-5474-8395-9879-0f7c2ab16d3d

## Repository
✅ implementują interfejsy z domain
✅ używają sqlc
✅ mapują modele DB ↔ domain
✅ NIE zawierają business logic

```
internal/
├── domain/
├── application/
└── infrastructure/
    ├── db/
    └── repository/
```
```
    └── repository/
        ├── public/
        │   ├── tenant_repository.go
        │   └── user_repository.go
        │
        └── tenant/
            └── todo_repository.go
```

6. Dlaczego repository używa sqlc?

Bo sqlc daje:

✅ type-safe SQL
✅ compile-time validation
✅ brak ORM magic
✅ świetny performance

Repository dodaje:

✅ mapping
✅ domain abstraction
✅ hexagonal boundary

Co NIE powinno być w repository

❌ transaction management
❌ business rules
❌ tenant resolving
❌ HTTP
❌ DTO validation

Gdzie transaction?
W application layer.
Np:
txManager.WithinTransaction(...)

### Repository factory - TODO
Lepsza wersja (production)

Z czasem zrób:

repositories factory

Np:

type Repositories struct {
	Todos todo.Repository
	Users user.Repository
}

### Domain
domain/todo/entity.go
```
internal/
│
├── domain/
│   ├── tenant/
│   ├── user/
│   └── todo/
│
├── application/
│   ├── command/
│   └── query/
```

Interfej repozytorium:
domain/todo/repository.go 

Implementacja repozytorium:
infrastructure/repository/tenant/todo_repository.go

### Command
application/command/create_todo.go

### Czy query side MUSI używać repository?

Nie.

W CQRS często:

COMMAND SIDE
handler
→ repository
→ domain
QUERY SIDE
handler
→ sqlc directly

To bardzo popularne.

13. Co ja polecam Tobie
COMMAND SIDE

Repository + DDD.

QUERY SIDE

sqlc directly.

Np:

type ListTodosHandler struct {
	q *querydb.Queries
}

bez repository.

To jest bardzo clean.

### Transacke w command side
a transakcje powinny być przede wszystkim:
command side
W CQRS:
write side → transaction
read side → zwykle bez tx

### TxManager nie powinien znać pakietu (tylko interfejs?)
TxManager powinien znać tylko:
commanddb

### Query side - bez transakcji
Query side NIE powinien być w transaction managerze.

Bo później:

cache
replicas
read models
analytics DB

i query side może używać innej bazy.

Np.:
```
type TodoQueries struct {
	q *querydb.Queries
}
q := querydb.New(pool)

```

### defer tx.Rollback()
defer tx.Rollback()
jest poprawne.
Bo:
po commit rollback zwróci noop/error
pgx to bezpiecznie ignoruje
To production-grade pattern.

### TxManager - walidacja nazwy schematu
validateSchemaName

### Co z testami? Jak testować repository?

tests/integration/repository/todo_repository_test.go

repozytoria powinieneś testować głównie:

integration tests

a NIE mockami.

To bardzo ważne.

Największe ryzyko błędów to:

SQL
migracje
constraints
tx
schema
search_path
mapping
pgx/sqlc integration

Mocki tego NIE sprawdzają.

2. Co testować integration testami?

✅ sqlc queries
✅ repositories
✅ migrations
✅ multitenancy isolation
✅ transactions
✅ constraints
✅ rollback
✅ tenant schema logic

5. Struktura testów

Polecam:
```
tests/
└── integration/
    │
    ├── repository/
    │   ├── tenant_repository_test.go
    │   ├── user_repository_test.go
    │   └── todo_repository_test.go
    │
    ├── migrations/
    │
    └── helpers/
```


13. Co ten test REALNIE sprawdza?

✅ postgres
✅ migrations
✅ schema-per-tenant
✅ search_path
✅ tx manager
✅ sqlc generated code
✅ repository mapping
✅ insert/select
✅ constraints

To OGROMNA wartość.


14. Czy robić unit testy repository?

Najczęściej:

❌ NIE

Bo mockowanie sqlc/pgx daje małą wartość.

W Application layer

Mockuj:

repository interfaces

16. SUPER ważny test

Musisz mieć:

Tenant isolation test
func TestTenantIsolation(...)

Scenariusz:

tenant_a -> create todo
tenant_b -> should not see it

To krytyczny test dla SaaS.

17. Kolejny ważny test
Rollback test
create todo
panic/error
verify rollback
18. Jeszcze ważniejszy test
Concurrent tenant test

Sprawdza:

search_path leakage

BARDZO ważne przy pgxpool.


# TODO - continue
https://chatgpt.com/g/g-p-6974db69d3dc819196dfb45bdb7bd10a-saas/c/69fcbef5-5474-8395-9879-0f7c2ab16d3d

# 2. Warstwa Application 
                                      
  Tak, internal/application/ to właściwe miejsce. Warstwa aplikacji orchestruje logikę: przyjmuje Command/Query, używa repozytorium (przez interfejs domenowy), zwraca wynik.
                                                                                                                                                                                              
  Dla CQRS typowy podział:     
  internal/application/                                                                                                                                                                       
    command/      
      create_todo.go      ← handler dla mutacji
      complete_todo.go
      delete_todo.go
    query/
      get_todo.go         ← handler dla odczytów


 Kluczowa zasada: Application layer zna tylko interfejsy domenowe (todo.Repository), nie konkretne implementacje z infrastructure.

 