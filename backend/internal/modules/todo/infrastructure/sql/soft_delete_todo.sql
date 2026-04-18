UPDATE todos
SET deleted_at = $1, updated_at = $1
WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL;
