CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    schema_name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    migration_status TEXT DEFAULT 'pending',
    migration_error TEXT,
    migration_updated_at TIMESTAMP
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL
);