-- name: ListAuthors :many
SELECT id, name, bio, created_at
FROM authors
ORDER BY name;

-- name: GetAuthor :one
SELECT id, name, bio, created_at
FROM authors
WHERE id = ?;

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
INSERT INTO authors (name, bio) VALUES (?, ?) RETURNING id, name, bio, created_at;

-- name: CreatePost :one
INSERT INTO posts (author_id, title, body, views)
VALUES (?, ?, ?, ?)
RETURNING id, author_id, title, body, views, created_at;

-- name: DeleteAllPosts :exec
-- metaquery: off
-- (Seed/cleanup only; no builder/metadata needed.)
DELETE FROM posts;

-- name: DeleteAllAuthors :exec
-- metaquery: off
DELETE FROM authors;
