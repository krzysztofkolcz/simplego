# Jak połączyć migracje oraz multitenancy z sqlc.
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
```
migracje i db w roocie - dobre dla małych projektów.
## Sqlc config
```
my-migrations/
├── sqlc.yaml
```
## connection.go - połączenie z bazą przez pgxpool.Pool
```
my-migrations/
├── internal/
│   ├── database/            # connection + tenant mgmt
│   │   ├── connection.go
```

## tenant.go - zarządzanie tenantami ??? Czy najlepsze miejsce? Myślę, że lepiej przez handler

## 
my-migrations/
├── internal/
│   ├── db/                  # sqlc generated
│   ├── database/            # connection + tenant mgmt