-- name: CreateFeedFollow :one
WITH inserted_feed_follows AS (
    INSERT INTO feed_follows (created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4
    )
    RETURNING *
)
SELECT inserted_feed_follows.*, u.name  as "user_name", f.name as "feed_name"
FROM inserted_feed_follows
INNER JOIN feed f
    ON inserted_feed_follows.feed_id = f.id
INNER JOIN users as u
    ON inserted_feed_follows.user_id = u.id;

-- name: GetFeedFollowsForUser :many
SELECT ff.*, u.name  as "user_name", f.name as "feed_name"
FROM feed_follows  ff
INNER JOIN  feed f
ON ff.feed_id = f.id
INNER JOIN users u
ON ff.user_id = u.id
WHERE u.name = $1;

-- name: DeleteFeedFollowsByIds :exec
DELETE FROM feed_follows
WHERE user_id = $1 and feed_id = $2;