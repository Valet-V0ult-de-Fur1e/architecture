SELECT id, user_id, title, description, status, priority, deleted_at, created_at, updated_at
FROM todos
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;
