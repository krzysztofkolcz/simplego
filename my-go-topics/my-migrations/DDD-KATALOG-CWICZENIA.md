# DDD — Bounded Context "Katalog" (ćwiczenia)

## Opis funkcjonalności

Dwie encje:
- **Półprodukt (Component)** — stan magazynowy: nazwa, opis, kod, producent, url_producenta, cena, waga, ilość, zdjęcie
- **Produkt (Product)** — nazwa, cena, opis, tagi, relacja wiele-do-wielu z Półproduktami (z ilością)

## Czy to osobny Bounded Context?

Tak — **BC "Katalog"** obok istniejących BC (`todo`, `tenant`, `user`).

Oba pojęcia (półprodukt, produkt, receptura) żyją w tym samym języku domenowym — jeden BC wystarczy.

Relacja wiele-do-wielu z ilością wymaga encji pośredniej. W DDD to nie jest zwykła tabela łącząca — to pełnoprawna encja: `RecipeItem` (lub `ProductComponent`).

---

## Drzewo plików

```
internal/
├── domain/
│   ├── tenant/               ← istniejący BC
│   ├── todo/                 ← istniejący BC
│   ├── user/                 ← istniejący BC
│   └── catalog/              ← NOWY BC
│       ├── component.go      # encja: Półprodukt (Component)
│       ├── product.go        # encja: Produkt + lista RecipeItem
│       ├── recipe_item.go    # encja pośrednia: product_id, component_id, quantity
│       ├── repository.go     # interfejsy: ComponentRepository, ProductRepository
│       ├── errors.go         # ErrComponentNotFound, ErrProductNotFound itd.
│       └── events.go         # ComponentCreated, ProductCreated itd.
│
├── application/
│   ├── command/
│   │   ├── create_todo.go        ← istniejące
│   │   ├── create_component.go   # NOWE
│   │   ├── update_component.go
│   │   ├── create_product.go
│   │   └── add_component_to_product.go
│   ├── query/
│   │   ├── get_todo.go           ← istniejące
│   │   ├── get_component.go      # NOWE
│   │   ├── list_components.go
│   │   ├── get_product.go
│   │   └── list_products.go
│   └── port/
│       ├── todo_usecases.go      ← istniejące
│       └── catalog_usecases.go   # NOWE — interfejsy use-case'ów katalogu
│
├── infrastructure/
│   ├── db/
│   │   ├── migrations/tenant/
│   │   │   ├── 001_init.up.sql        ← istniejące (todos, outbox_events)
│   │   │   ├── 002_outbox.up.sql      ← istniejące
│   │   │   ├── 003_catalog.up.sql     # NOWE
│   │   │   └── 003_catalog.down.sql
│   │   ├── queries/
│   │   │   ├── command/tenant/
│   │   │   │   ├── todos.sql          ← istniejące
│   │   │   │   └── catalog.sql        # NOWE
│   │   │   └── query/tenant/
│   │   │       ├── todos.sql          ← istniejące
│   │   │       └── catalog.sql        # NOWE
│   │   └── sqlc/                      # wygenerowane — make sqlc-generate
│   │
│   ├── repository/
│   │   ├── tenant/               ← istniejące
│   │   └── catalog/              # NOWE
│   │       ├── component_repository.go
│   │       └── product_repository.go
│   │
│   └── usecase/
│       ├── todo_usecases.go      ← istniejące
│       └── catalog_usecases.go   # NOWE — wiring: handler + repo + publisher
│
└── http/
    ├── handler/
    │   ├── server.go             # dodać pola: CreateComponent, GetProduct itd.
    │   ├── todo_handler.go       ← istniejące
    │   └── catalog_handler.go   # NOWE
    └── router/
        └── router.go             # dodać endpointy /v1/components, /v1/products
```

---

## Migracja SQL (`003_catalog.up.sql`)

```sql
CREATE TABLE components (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL,
    description      TEXT,
    code             TEXT NOT NULL UNIQUE,
    manufacturer     TEXT,
    manufacturer_url TEXT,
    price            NUMERIC(12,2),
    weight_g         INTEGER,
    quantity         INTEGER NOT NULL DEFAULT 0,
    image_url        TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    price       NUMERIC(12,2),
    tags        TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE product_components (
    product_id   UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    component_id UUID NOT NULL REFERENCES components(id) ON DELETE RESTRICT,
    quantity     INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (product_id, component_id)
);
```

---

## Decyzje projektowe

| Kwestia | Decyzja |
|---|---|
| Osobny BC? | Tak — `catalog/` obok `todo/`, `tenant/`, `user/` |
| Encja pośrednia | `RecipeItem` w domenie, tabela `product_components` w DB |
| `tags` | `TEXT[]` w PostgreSQL — nie osobna tabela, chyba że potrzebne filtrowanie z indeksem GIN |
| Zdjęcie | `image_url TEXT` — URL do S3/CDN, nie blob w bazie |
| Schemat | Tenant — każdy tenant ma swój katalog (analogicznie do todos) |
| Po zmianie queries/migracji | `make sqlc-generate` |
