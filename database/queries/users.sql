-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateUserGoogle :one
INSERT INTO users (first_name, last_name, email, google_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: LinkUserGoogleIDByEmail :one
UPDATE users
SET google_id = $2, updated_at = NOW()
WHERE email = $1 AND google_id IS NULL
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
  first_name = $2,
  last_name = $3,
  email = $4,
  password_hash = $5,
  monthly_income = $6
WHERE id = $1
RETURNING *;
