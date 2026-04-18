UPDATE todos
SET status = $1, updated_at = $2
WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL;
