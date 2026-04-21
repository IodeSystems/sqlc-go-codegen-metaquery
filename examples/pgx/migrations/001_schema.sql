CREATE TABLE authors (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    bio        TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE posts (
    id         BIGSERIAL PRIMARY KEY,
    author_id  BIGINT      NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    title      TEXT        NOT NULL,
    body       TEXT,
    views      BIGINT      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX posts_author_id_idx ON posts(author_id);
