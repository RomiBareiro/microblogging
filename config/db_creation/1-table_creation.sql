CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL CHECK (char_length(content) <= 280),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE CHECK (char_length(username) <= 20)
);

CREATE TABLE IF NOT EXISTS follows (
    follower_id UUID NOT NULL,
    followee_id UUID NOT NULL,
    PRIMARY KEY (follower_id, followee_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (followee_id) REFERENCES users(id) ON DELETE CASCADE
);
