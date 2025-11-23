-- name: GetMessageReaction :one
SELECT * FROM message_reactions
WHERE message_id = $1 AND user_id = $2 LIMIT 1;

-- name: GetMessageReactionWithDetails :one
SELECT mr.*, u.phone_number, u.display_name
FROM message_reactions mr
JOIN users u ON mr.user_id = u.id
WHERE mr.message_id = $1 AND mr.user_id = $2 LIMIT 1;

-- name: ListMessageReactions :many
SELECT * FROM message_reactions
WHERE message_id = $1
ORDER BY reacted_at ASC;

-- name: ListMessageReactionsWithDetails :many
SELECT mr.*, u.phone_number, u.display_name
FROM message_reactions mr
JOIN users u ON mr.user_id = u.id
WHERE mr.message_id = $1
ORDER BY mr.reacted_at ASC;

-- name: ListUserReactions :many
SELECT mr.*, m.content as message_content, c.title as conversation_title
FROM message_reactions mr
JOIN messages m ON mr.message_id = m.id
JOIN conversations c ON m.conversation_id = c.id
WHERE mr.user_id = $1
ORDER BY mr.reacted_at DESC;

-- name: AddMessageReaction :one
INSERT INTO message_reactions (
  message_id, user_id, reaction
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateMessageReaction :one
UPDATE message_reactions
SET reaction = $3,
    reacted_at = NOW()
WHERE message_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveMessageReaction :exec
DELETE FROM message_reactions
WHERE message_id = $1 AND user_id = $2;

-- name: RemoveAllMessageReactions :exec
DELETE FROM message_reactions
WHERE message_id = $1;

-- name: CountMessageReactions :one
SELECT COUNT(*) FROM message_reactions
WHERE message_id = $1;

-- name: CountMessageReactionsByType :one
SELECT COUNT(*) FROM message_reactions
WHERE message_id = $1 AND reaction = $2;

-- name: GetMessageReactionSummary :many
SELECT reaction, COUNT(*) as count
FROM message_reactions
WHERE message_id = $1
GROUP BY reaction
ORDER BY count DESC;

-- name: GetUserReactionOnMessage :one
SELECT reaction FROM message_reactions
WHERE message_id = $1 AND user_id = $2 LIMIT 1;

-- name: HasUserReactedToMessage :one
SELECT EXISTS(
  SELECT 1 FROM message_reactions
  WHERE message_id = $1 AND user_id = $2
);

-- name: ListMessagesByReaction :many
SELECT DISTINCT m.*, c.title as conversation_title
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
JOIN message_reactions mr ON m.id = mr.message_id
WHERE mr.reaction = $1
ORDER BY m.created_at DESC
LIMIT 50;