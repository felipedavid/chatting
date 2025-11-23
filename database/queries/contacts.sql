-- name: GetContact :one
SELECT * FROM contacts
WHERE user_id = $1 AND contact_id = $2 LIMIT 1;

-- name: GetContactWithDetails :one
SELECT c.*, u.phone_number as contact_phone, u.display_name as contact_display_name
FROM contacts c
JOIN users u ON c.contact_id = u.id
WHERE c.user_id = $1 AND c.contact_id = $2 LIMIT 1;

-- name: ListUserContacts :many
SELECT * FROM contacts
WHERE user_id = $1
ORDER BY contact_name ASC;

-- name: ListUserContactsWithDetails :many
SELECT c.*, u.phone_number as contact_phone, u.display_name as contact_display_name, u.about as contact_about
FROM contacts c
JOIN users u ON c.contact_id = u.id
WHERE c.user_id = $1
ORDER BY c.contact_name ASC;

-- name: ListContactUsers :many
SELECT * FROM contacts
WHERE contact_id = $1
ORDER BY contact_name ASC;

-- name: AddContact :one
INSERT INTO contacts (
  user_id, contact_id, contact_name
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateContactName :one
UPDATE contacts
SET contact_name = $3
WHERE user_id = $1 AND contact_id = $2
RETURNING *;

-- name: RemoveContact :exec
DELETE FROM contacts
WHERE user_id = $1 AND contact_id = $2;

-- name: RemoveAllUserContacts :exec
DELETE FROM contacts
WHERE user_id = $1;

-- name: CountUserContacts :one
SELECT COUNT(*) FROM contacts
WHERE user_id = $1;

-- name: SearchUserContacts :many
SELECT c.*, u.phone_number as contact_phone, u.display_name as contact_display_name
FROM contacts c
JOIN users u ON c.contact_id = u.id
WHERE c.user_id = $1 
  AND (c.contact_name ILIKE '%' || $2 || '%' 
       OR u.phone_number LIKE $2 || '%'
       OR u.display_name ILIKE '%' || $2 || '%')
ORDER BY c.contact_name ASC
LIMIT 20;

-- name: IsUserContact :one
SELECT EXISTS(
  SELECT 1 FROM contacts
  WHERE user_id = $1 AND contact_id = $2
);

-- name: GetMutualContacts :many
SELECT c1.contact_id, c1.contact_name as user1_name, c2.contact_name as user2_name,
       u.phone_number, u.display_name
FROM contacts c1
JOIN contacts c2 ON c1.contact_id = c2.user_id AND c2.contact_id = c1.user_id
JOIN users u ON c1.contact_id = u.id
WHERE c1.user_id = $1
ORDER BY c1.contact_name ASC;

-- name: GetCommonContacts :many
SELECT c.*, u.phone_number as contact_phone, u.display_name as contact_display_name
FROM contacts c
JOIN users u ON c.contact_id = u.id
WHERE c.user_id = $1
  AND c.contact_id IN (
    SELECT contact_id FROM contacts c2 WHERE c2.user_id = $2
  )
ORDER BY c.contact_name ASC;

-- name: ListContactSuggestions :many
SELECT u.*, 'mutual' as suggestion_type
FROM users u
WHERE u.id IN (
  SELECT contact_id FROM contacts c2 WHERE c2.user_id IN (
    SELECT contact_id FROM contacts c3 WHERE c3.user_id = $1
  ) AND c2.contact_id != $1
  AND c2.contact_id NOT IN (SELECT contact_id FROM contacts c4 WHERE c4.user_id = $1)
)
LIMIT 10;