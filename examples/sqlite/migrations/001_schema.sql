CREATE TABLE authors (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT     NOT NULL,
    bio        TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    author_id  INTEGER  NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    title      TEXT     NOT NULL,
    body       TEXT,
    views      INTEGER  NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX posts_author_id_idx ON posts(author_id);
