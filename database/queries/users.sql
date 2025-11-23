-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByPhoneNumber :one
SELECT * FROM users
WHERE phone_number = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY display_name, phone_number;

-- name: CreateUser :one
INSERT INTO users (
  phone_number, display_name, about
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET phone_number = $2,
    display_name = $3,
    about = $4
WHERE id = $1
RETURNING *;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserAbout :one
UPDATE users
SET about = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: SearchUsersByPhone :many
SELECT * FROM users
WHERE phone_number LIKE $1 || '%'
ORDER BY phone_number
LIMIT 20;

-- name: SearchUsersByDisplayName :many
SELECT * FROM users
WHERE display_name ILIKE '%' || $1 || '%'
ORDER BY display_name
LIMIT 20;