CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- todo_status mirrors domain.Status exactly.
-- Using a Postgres ENUM type enforces valid values at the DB level,
-- which complements the domain-level invariant in the aggregate.
CREATE TYPE todo_status AS ENUM ('pending', 'done');

CREATE TABLE todos (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT        NOT NULL CHECK (length(title) > 0),
    description TEXT,
    status      todo_status NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
