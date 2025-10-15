-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, user_id, url)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1 LIMIT 1;