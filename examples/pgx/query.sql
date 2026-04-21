-- name: ListAuthors :many
SELECT id, name, bio, created_at
FROM authors
ORDER BY name;

-- name: GetAuthor :one
SELECT id, name, bio, created_at
FROM authors
WHERE id = $1;

-- name: ListPosts :many
SELECT id, author_id, title, body, views, created_at
FROM posts
ORDER BY created_at DESC;

-- name: ListPostsWithAuthor :many
SELECT p.id, p.author_id, p.title, p.body, p.views, p.created_at, a.name AS author_name
FROM posts p
JOIN authors a ON a.id = p.author_id
ORDER BY p.created_at DESC;

-- name: CreateAuthor :one
INSERT INTO authors (name, bio) VALUES ($1, $2) RETURNING id, name, bio, created_at;

-- name: CreatePost :one
INSERT INTO posts (author_id, title, body, views)
VALUES ($1, $2, $3, $4)
RETURNING id, author_id, title, body, views, created_at;

-- name: TruncateAll :exec
-- metaquery: off
-- (TruncateAll is only invoked from seed code; no builder/metadata needed.)
TRUNCATE authors RESTART IDENTITY CASCADE;
