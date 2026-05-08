-- name: GetTenantByID :one
SELECT
    id,
    schema_name,
    migration_status,
    migration_error,
    migration_updated_at,
    created_at
FROM tenants
WHERE id = $1;

-- name: GetTenantBySchema :one
SELECT
    id,
    schema_name,
    migration_status,
    migration_error,
    migration_updated_at,
    created_at
FROM tenants
WHERE schema_name = $1;