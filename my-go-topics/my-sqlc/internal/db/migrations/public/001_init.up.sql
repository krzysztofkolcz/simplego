-- 001_init.up.sql

CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    schema_name TEXT UNIQUE NOT NULL
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL
);