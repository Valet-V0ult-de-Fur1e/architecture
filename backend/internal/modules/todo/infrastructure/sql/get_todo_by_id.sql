SELECT id, user_id, title, description, status, priority, deleted_at, created_at, updated_at
FROM todos
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
