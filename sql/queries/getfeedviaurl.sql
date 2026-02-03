-- name: GetFeedViaUrl :one
SELECT * 
FROM feeds
WHERE url = $1;