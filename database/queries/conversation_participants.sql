-- name: GetConversationParticipant :one
SELECT * FROM conversation_participants
WHERE conversation_id = $1 AND user_id = $2 LIMIT 1;

-- name: GetConversationParticipantWithDetails :one
SELECT cp.*, u.phone_number, u.display_name
FROM conversation_participants cp
JOIN users u ON cp.user_id = u.id
WHERE cp.conversation_id = $1 AND cp.user_id = $2 LIMIT 1;

-- name: ListConversationParticipants :many
SELECT * FROM conversation_participants
WHERE conversation_id = $1
ORDER BY joined_at ASC;

-- name: ListConversationParticipantsWithDetails :many
SELECT cp.*, u.phone_number, u.display_name
FROM conversation_participants cp
JOIN users u ON cp.user_id = u.id
WHERE cp.conversation_id = $1
ORDER BY cp.joined_at ASC;

-- name: ListUserConversations :many
SELECT cp.*, c.title, c.is_group, c.created_at as conversation_created_at
FROM conversation_participants cp
JOIN conversations c ON cp.conversation_id = c.id
WHERE cp.user_id = $1
ORDER BY cp.joined_at DESC;

-- name: ListUserConversationsWithLastMessage :many
SELECT cp.*, c.title, c.is_group, c.created_at as conversation_created_at,
       m.content as last_message_content, m.created_at as last_message_at
FROM conversation_participants cp
JOIN conversations c ON cp.conversation_id = c.id
LEFT JOIN messages m ON m.conversation_id = c.id
WHERE cp.user_id = $1
  AND m.created_at = (
    SELECT MAX(created_at) 
    FROM messages 
    WHERE conversation_id = c.id
  )
ORDER BY m.created_at DESC;

-- name: AddConversationParticipant :one
INSERT INTO conversation_participants (
  conversation_id, user_id, role
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateParticipantRole :one
UPDATE conversation_participants
SET role = $3
WHERE conversation_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveConversationParticipant :exec
DELETE FROM conversation_participants
WHERE conversation_id = $1 AND user_id = $2;

-- name: RemoveAllConversationParticipants :exec
DELETE FROM conversation_participants
WHERE conversation_id = $1;

-- name: CountConversationParticipants :one
SELECT COUNT(*) FROM conversation_participants
WHERE conversation_id = $1;

-- name: IsUserInConversation :one
SELECT EXISTS(
  SELECT 1 FROM conversation_participants
  WHERE conversation_id = $1 AND user_id = $2
);