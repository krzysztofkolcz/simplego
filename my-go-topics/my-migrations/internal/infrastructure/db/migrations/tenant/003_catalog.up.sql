CREATE TABLE components (
    id               UUID PRIMARY KEY,
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
    id          UUID PRIMARY KEY,
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