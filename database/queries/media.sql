-- name: GetMedia :one
SELECT * FROM media
WHERE id = $1 LIMIT 1;

-- name: GetMediaByMessage :many
SELECT * FROM media
WHERE message_id = $1;

-- name: ListMedia :many
SELECT * FROM media
ORDER BY uploaded_at DESC
LIMIT $1 OFFSET $2;

-- name: ListMediaByMimeType :many
SELECT * FROM media
WHERE mime_type = $1
ORDER BY uploaded_at DESC;

-- name: ListMediaByMimeTypePrefix :many
SELECT * FROM media
WHERE mime_type LIKE $1 || '%'
ORDER BY uploaded_at DESC;

-- name: CreateMedia :one
INSERT INTO media (
  message_id, file_url, mime_type, file_size
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateMediaFileInfo :one
UPDATE media
SET file_url = $2,
    file_size = $3
WHERE id = $1
RETURNING *;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE id = $1;

-- name: DeleteMediaByMessage :exec
DELETE FROM media
WHERE message_id = $1;

-- name: CountMediaByMessage :one
SELECT COUNT(*) FROM media
WHERE message_id = $1;

-- name: GetMediaByFileUrl :one
SELECT * FROM media
WHERE file_url = $1 LIMIT 1;

-- name: ListMediaBySizeRange :many
SELECT * FROM media
WHERE file_size >= $1 AND file_size <= $2
ORDER BY file_size ASC;

-- name: GetTotalMediaSize :one
SELECT COALESCE(SUM(file_size), 0) FROM media;

-- name: GetMediaSizeByMessage :one
SELECT COALESCE(SUM(file_size), 0) FROM media
WHERE message_id = $1;

-- name: ListMediaByUploadTimeRange :many
SELECT * FROM media
WHERE uploaded_at >= $1 AND uploaded_at <= $2
ORDER BY uploaded_at DESC;