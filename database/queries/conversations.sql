-- name: GetConversation :one
SELECT * FROM conversations
WHERE id = $1 LIMIT 1;

-- name: GetConversationByIdWithCreator :one
SELECT c.*, u.phone_number as creator_phone, u.display_name as creator_name
FROM conversations c
LEFT JOIN users u ON c.created_by = u.id
WHERE c.id = $1 LIMIT 1;

-- name: ListConversations :many
SELECT * FROM conversations
ORDER BY created_at DESC;

-- name: ListConversationsByCreator :many
SELECT * FROM conversations
WHERE created_by = $1
ORDER BY created_at DESC;

-- name: ListGroupConversations :many
SELECT * FROM conversations
WHERE is_group = true
ORDER BY created_at DESC;

-- name: CreateConversation :one
INSERT INTO conversations (
  is_group, title, created_by
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateConversation :one
UPDATE conversations
SET title = $2
WHERE id = $1
RETURNING *;

-- name: UpdateConversationTitle :one
UPDATE conversations
SET title = $2
WHERE id = $1
RETURNING *;

-- name: DeleteConversation :exec
DELETE FROM conversations
WHERE id = $1;

-- name: SearchConversationsByTitle :many
SELECT * FROM conversations
WHERE title ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT 20;