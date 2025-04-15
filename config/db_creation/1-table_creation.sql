CREATE TABLE posts (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL CHECK (char_length(content) <= 280),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE follows (
    follower_id TEXT NOT NULL,
    following_id TEXT NOT NULL,
    PRIMARY KEY (follower_id, following_id)
);
