```
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Struktura projektu:
```
project/
├── db/
│   ├── schema.sql
│   ├── query.sql
├── internal/
│   └── db/
│       └── (tu wygeneruje się kod)
├── main.go
└── sqlc.yaml  # sqlc.yaml na tym poziomie
```

```
sqlc generate
```
Wygenerowanie na podstawie sqlc.yaml
./db/schema.go
./db/query.go

Architektura:
```
my-sqlc/
├── db/
│   ├── schema.sql
│   ├── query.sql
├── internal/
│   ├── application/
│   └── db/
│       └── (tu wygeneruje się kod)
│   ├── domain
│   ├── infrastructure/ 
│       └── persistence
│           └── user_repository.go 
├── tests/
|   ├── db_test.go
|   ├── test_main.go   👈 start kontenera, migracje, 
|   ├── helpers.go
├── main.go
└── sqlc.yaml  # sqlc.yaml na tym poziomie

```

todo - singletone container

# Migracje. Multitenancy + schema per tenant
## golang-migrate
Struktura migracji:
```
db/
├── migrations/
│   ├── public/
│   │   ├── 001_init.up.sql
│   │   └── 001_init.down.sql
│   │
│   └── tenant/
│       ├── 001_init.up.sql
│       └── 001_init.down.sql
```

Uruchomienie migracji public:
```
migrate -path db/migrations/public \
  -database "postgres://..." up
```

TENANT (dla każdego schematu)
-stworzyć schema
-ustawić search_path
-odpalić migracje
```
my-sqlc/
├── internal/
│   └── tenant/
```


TODO:
👉 to jest boilerplate, ale w produkcji dodasz:

retry migracji
worker pool dla tenantów
observability (OpenTelemetry)
outbox pattern
Stripe billing per tenant

🔥 worker pool do migracji tenantów (ważne przy 1000+)
💳 integrację Stripe + tenant schema
🧪 testy multitenancy (izolacja danych)
📦 event-driven (outbox + kolejka zamiast RabbitMQ)

### Struktura DDD
```
project/
├── cmd/
│   └── api/
│       └── main.go

├── internal/
│   ├── domain/                     # 🔴 PURE DOMAIN (DDD)
│   │   └── todo/
│   │       ├── entity.go
│   │       ├── errors.go
│   │       └── repository.go       # port
│   │
│   ├── application/                # 🟡 USE CASES (CQRS)
│   │   └── todo/
│   │       ├── command/
│   │       │   ├── create_todo.go
│   │       │   └── complete_todo.go
│   │       │
│   │       └── query/
│   │           └── list_todos.go
│   │
│   ├── infrastructure/             # 🔵 ADAPTERY (hexagonal)
│   │   ├── persistence/
│   │   │   ├── db.go               # pgxpool
│   │   │   ├── tx.go               # transaction manager
│   │   │   └── todo_repository.go  # implementacja portu
│   │   │
│   │   └── http/
│   │       └── handler.go
│   │
│   ├── db/                         # sqlc generated
│   │
│   └── bootstrap/
│       └── wiring.go               # DI (składanie systemu)

├── db/
│   ├── migrations/
│   │   ├── 001_init.up.sql
│   │   └── 001_init.down.sql
│   │
│   ├── query/
│   │   └── todo.sql
│   │
│   └── sqlc.yaml

├── tests/
│   ├── integration/
│   │   ├── test_main.go            # testcontainers
│   │   └── todo_test.go
│   │
│   └── unit/
│       └── todo_service_test.go

├── go.mod
└── Makefile
```