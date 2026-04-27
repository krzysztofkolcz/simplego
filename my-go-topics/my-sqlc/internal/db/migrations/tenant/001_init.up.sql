-- tenant/001_init.up.sql

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    amount INT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);