-- name: GetMessageReceipt :one
SELECT * FROM message_receipts
WHERE message_id = $1 AND user_id = $2 LIMIT 1;

-- name: GetMessageReceiptWithDetails :one
SELECT mr.*, u.phone_number, u.display_name
FROM message_receipts mr
JOIN users u ON mr.user_id = u.id
WHERE mr.message_id = $1 AND mr.user_id = $2 LIMIT 1;

-- name: ListMessageReceipts :many
SELECT * FROM message_receipts
WHERE message_id = $1
ORDER BY delivered_at ASC, read_at ASC;

-- name: ListMessageReceiptsWithDetails :many
SELECT mr.*, u.phone_number, u.display_name
FROM message_receipts mr
JOIN users u ON mr.user_id = u.id
WHERE mr.message_id = $1
ORDER BY mr.delivered_at ASC, mr.read_at ASC;

-- name: ListUserMessageReceipts :many
SELECT mr.*, m.content as message_content, m.created_at as message_created_at
FROM message_receipts mr
JOIN messages m ON mr.message_id = m.id
WHERE mr.user_id = $1
ORDER BY m.created_at DESC;

-- name: ListUnreadMessages :many
SELECT m.*, c.title as conversation_title
FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE m.id NOT IN (
  SELECT message_id FROM message_receipts 
  WHERE user_id = $1 AND read_at IS NOT NULL
) AND m.sender_id != $1
ORDER BY m.created_at DESC;

-- name: CreateMessageReceipt :one
INSERT INTO message_receipts (
  message_id, user_id, delivered_at, read_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: MarkMessageAsDelivered :one
INSERT INTO message_receipts (
  message_id, user_id, delivered_at
) VALUES (
  $1, $2, NOW()
)
ON CONFLICT (message_id, user_id) 
DO UPDATE SET delivered_at = NOW()
RETURNING *;

-- name: MarkMessageAsRead :one
INSERT INTO message_receipts (
  message_id, user_id, delivered_at, read_at
) VALUES (
  $1, $2, NOW(), NOW()
)
ON CONFLICT (message_id, user_id) 
DO UPDATE SET read_at = NOW()
RETURNING *;

-- name: UpdateMessageReceipt :one
UPDATE message_receipts
SET delivered_at = $3,
    read_at = $4
WHERE message_id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteMessageReceipt :exec
DELETE FROM message_receipts
WHERE message_id = $1 AND user_id = $2;

-- name: DeleteAllMessageReceipts :exec
DELETE FROM message_receipts
WHERE message_id = $1;

-- name: CountMessageDeliveries :one
SELECT COUNT(*) FROM message_receipts
WHERE message_id = $1 AND delivered_at IS NOT NULL;

-- name: CountMessageReads :one
SELECT COUNT(*) FROM message_receipts
WHERE message_id = $1 AND read_at IS NOT NULL;

-- name: GetMessageDeliveryStatus :one
SELECT 
  COUNT(*) as total_recipients,
  COUNT(delivered_at) as delivered_count,
  COUNT(read_at) as read_count
FROM message_receipts
WHERE message_id = $1;

-- name: GetUnreadMessageCount :one
SELECT COUNT(*) FROM messages m
WHERE m.conversation_id = $1 
  AND m.sender_id != $2
  AND m.id NOT IN (
    SELECT message_id FROM message_receipts 
    WHERE user_id = $2 AND read_at IS NOT NULL
  );