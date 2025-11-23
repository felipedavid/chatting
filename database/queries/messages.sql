-- name: GetMessage :one
SELECT * FROM messages
WHERE id = $1 LIMIT 1;

-- name: GetMessageWithDetails :one
SELECT m.*, u.phone_number as sender_phone, u.display_name as sender_name
FROM messages m
LEFT JOIN users u ON m.sender_id = u.id
WHERE m.id = $1 LIMIT 1;

-- name: ListMessages :many
SELECT * FROM messages
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListConversationMessages :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: ListConversationMessagesWithDetails :many
SELECT m.*, u.phone_number as sender_phone, u.display_name as sender_name
FROM messages m
LEFT JOIN users u ON m.sender_id = u.id
WHERE m.conversation_id = $1
ORDER BY m.created_at ASC;

-- name: ListConversationMessagesPaginated :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUserMessages :many
SELECT m.*, c.title as conversation_title, c.is_group
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE m.sender_id = $1
ORDER BY m.created_at DESC;

-- name: ListReplyMessages :many
SELECT * FROM messages
WHERE reply_to_id = $1
ORDER BY created_at ASC;

-- name: CreateMessage :one
INSERT INTO messages (
  conversation_id, sender_id, content, message_type, reply_to_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateMessageContent :one
UPDATE messages
SET content = $2
WHERE id = $1
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1;

-- name: DeleteConversationMessages :exec
DELETE FROM messages
WHERE conversation_id = $1;

-- name: CountConversationMessages :one
SELECT COUNT(*) FROM messages
WHERE conversation_id = $1;

-- name: CountUserMessages :one
SELECT COUNT(*) FROM messages
WHERE sender_id = $1;

-- name: SearchMessages :many
SELECT m.*, c.title as conversation_title
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE m.content ILIKE '%' || $1 || '%'
ORDER BY m.created_at DESC
LIMIT 50;

-- name: GetLatestConversationMessage :one
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: GetMessagesAfterTimestamp :many
SELECT * FROM messages
WHERE conversation_id = $1 AND created_at > $2
ORDER BY created_at ASC;

-- name: GetMessagesBeforeTimestamp :many
SELECT * FROM messages
WHERE conversation_id = $1 AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;