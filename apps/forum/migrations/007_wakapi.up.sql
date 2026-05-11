CREATE TABLE user_wakapi (
    user_id      UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    instance_url TEXT NOT NULL,       -- e.g. https://waka.honeygarden.space
    api_key      TEXT NOT NULL,
    username     TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
