-- name: CreateFeed :one
INSERT INTO feed (id, created_at, updated_at, last_fetched_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    NULL,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: ListFeeds :many
SELECT f.*, u.name as "user_name"
FROM feed f
INNER JOIN users u
ON f.user_id = u.id;

-- name: GetFeedFromUrl :one
SELECT f.*
FROM feed f
INNER JOIN users u
ON f.user_id = u.id
where f.url = $1;

-- name: MarkFeedFetched :exec
UPDATE feed
SET updated_at = CURRENT_TIMESTAMP, last_fetched_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feed
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;