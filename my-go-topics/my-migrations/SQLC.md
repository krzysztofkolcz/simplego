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