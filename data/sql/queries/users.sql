-- name: CreateUser :one
INSERT INTO users (
    login,
    password,
    last_name,
    first_name,
    middle_name,
    role,
    auth_provider,
    language,
    notification_language,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByLogin :one
SELECT * FROM users
WHERE login = $1
LIMIT 1;

-- name: ListUsers :many
SELECT
    *,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    (
        $1::text = ''
        OR login ILIKE '%' || $1 || '%'
        OR (last_name || ' ' || first_name || ' ' || middle_name) ILIKE '%' || $1 || '%'
        OR (first_name || ' ' || last_name) ILIKE '%' || $1 || '%'
    )
    AND ($2::user_role IS NULL OR role = $2)
    AND ($3::boolean IS NULL OR active = $3)
    AND ($4::boolean IS NULL OR registered = $4)
ORDER BY
    CASE WHEN $5::text = 'login'      THEN login      END ASC,
    CASE WHEN $5::text = 'last_name'  THEN last_name  END ASC,
    CASE WHEN $5::text = 'created_at' THEN created_at::text END ASC,
    CASE WHEN $5::text = 'last_online' THEN last_online::text END DESC NULLS LAST,
    created_at DESC
LIMIT  $6
OFFSET $7;

-- name: UpdateUserProfile :one
UPDATE users SET
    last_name             = COALESCE(sqlc.narg(last_name),             last_name),
    first_name            = COALESCE(sqlc.narg(first_name),            first_name),
    middle_name           = COALESCE(sqlc.narg(middle_name),           middle_name),
    language              = COALESCE(sqlc.narg(language),              language),
    notification_language = COALESCE(sqlc.narg(notification_language), notification_language),
    updated_by            = sqlc.narg(updated_by),
    updated_at            = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserCredentials :one
UPDATE users SET
    login      = COALESCE(sqlc.narg(login),    login),
    password   = COALESCE(sqlc.narg(password), password),
    updated_by = sqlc.narg(updated_by),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserRole :one
UPDATE users SET
    role       = $2,
    updated_by = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users SET
    active     = COALESCE(sqlc.narg(active),     active),
    registered = COALESCE(sqlc.narg(registered), registered),
    updated_by = sqlc.narg(updated_by),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserLastOnline :exec
UPDATE users SET
    last_online = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: SoftDeleteUser :one
UPDATE users SET
    active     = FALSE,
    updated_by = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
