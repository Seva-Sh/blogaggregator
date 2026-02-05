-- name: MarkFeedFetched :one
UPDATE feeds
SET updated_at = CURRENT_TIMESTAMP,
    last_fetched_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;