-- name: GetEncryptionKey :one
SELECT * FROM encryption_keys
WHERE device_id = $1 LIMIT 1;

-- name: GetEncryptionKeyWithDevice :one
SELECT ek.*, ud.device_name, ud.device_type, ud.public_key, ud.user_id
FROM encryption_keys ek
JOIN user_devices ud ON ek.device_id = ud.id
WHERE ek.device_id = $1 LIMIT 1;

-- name: ListEncryptionKeys :many
SELECT * FROM encryption_keys
ORDER BY device_id;

-- name: ListEncryptionKeysByUser :many
SELECT ek.*, ud.device_name, ud.device_type
FROM encryption_keys ek
JOIN user_devices ud ON ek.device_id = ud.id
WHERE ud.user_id = $1
ORDER BY ek.device_id;

-- name: ListEncryptionKeysWithDeviceInfo :many
SELECT ek.*, ud.device_name, ud.device_type, ud.public_key, 
       u.phone_number, u.display_name
FROM encryption_keys ek
JOIN user_devices ud ON ek.device_id = ud.id
JOIN users u ON ud.user_id = u.id
ORDER BY u.phone_number, ud.device_name;

-- name: CreateEncryptionKey :one
INSERT INTO encryption_keys (
  device_id, identity_key, prekey_public, prekey_signature
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateEncryptionKey :one
UPDATE encryption_keys
SET identity_key = $2,
    prekey_public = $3,
    prekey_signature = $4
WHERE device_id = $1
RETURNING *;

-- name: UpdateIdentityKey :one
UPDATE encryption_keys
SET identity_key = $2
WHERE device_id = $1
RETURNING *;

-- name: UpdatePrekey :one
UPDATE encryption_keys
SET prekey_public = $2,
    prekey_signature = $3
WHERE device_id = $1
RETURNING *;

-- name: DeleteEncryptionKey :exec
DELETE FROM encryption_keys
WHERE device_id = $1;

-- name: DeleteEncryptionKeysByUser :exec
DELETE FROM encryption_keys
WHERE device_id IN (
  SELECT id FROM user_devices WHERE user_id = $1
);

-- name: CountUserEncryptionKeys :one
SELECT COUNT(*) FROM encryption_keys ek
JOIN user_devices ud ON ek.device_id = ud.id
WHERE ud.user_id = $1;

-- name: GetDeviceEncryptionKey :one
SELECT * FROM encryption_keys
WHERE device_id = $1 LIMIT 1;

-- name: ValidatePrekeySignature :one
SELECT EXISTS(
  SELECT 1 FROM encryption_keys
  WHERE device_id = $1 AND prekey_signature = $2
);

-- name: ListDevicesNeedingKeyRefresh :many
SELECT ud.*, u.phone_number, u.display_name
FROM user_devices ud
JOIN users u ON ud.user_id = u.id
WHERE ud.id NOT IN (SELECT device_id FROM encryption_keys)
ORDER BY ud.created_at DESC;