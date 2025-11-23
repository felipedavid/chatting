-- name: GetUserDevice :one
SELECT * FROM user_devices
WHERE id = $1 LIMIT 1;

-- name: GetUserDeviceByUserAndKey :one
SELECT * FROM user_devices
WHERE user_id = $1 AND public_key = $2 LIMIT 1;

-- name: ListUserDevices :many
SELECT * FROM user_devices
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListAllUserDevices :many
SELECT ud.*, u.phone_number, u.display_name
FROM user_devices ud
JOIN users u ON ud.user_id = u.id
ORDER BY ud.created_at DESC;

-- name: CreateUserDevice :one
INSERT INTO user_devices (
  user_id, device_name, device_type, public_key
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateUserDevice :one
UPDATE user_devices
SET device_name = $2,
    device_type = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUserDevice :exec
DELETE FROM user_devices
WHERE id = $1;

-- name: DeleteUserDevicesByUser :exec
DELETE FROM user_devices
WHERE user_id = $1;

-- name: CountUserDevices :one
SELECT COUNT(*) FROM user_devices
WHERE user_id = $1;