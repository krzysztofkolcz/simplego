-- name: CreateTenant :one
INSERT INTO tenants (
    id,
    schema_name
)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: UpdateTenantMigrationStatus :exec
UPDATE tenants
SET
    migration_status = $2,
    migration_error = $3,
    migration_updated_at = NOW()
WHERE id = $1;