CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    schema_name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL
);