-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, monthly_income)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users SET name = $2, email = $3, password_hash = $4, monthly_income = $5 WHERE id = $1 RETURNING *;