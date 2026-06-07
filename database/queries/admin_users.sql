-- name: SearchUsersByEmail :many
SELECT id, email, first_name, last_name, created_at
FROM users
WHERE email ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2;

-- name: ListRecentUsers :many
SELECT id, email, first_name, last_name, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1;

-- name: GetUserByIDAdmin :one
SELECT id, email, first_name, last_name, created_at
FROM users
WHERE id = $1
LIMIT 1;
