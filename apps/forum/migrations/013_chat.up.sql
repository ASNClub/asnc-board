-- Общий чат форума: одна публичная стена, без комнат и DM.
CREATE TABLE chat_messages (
    id         UUID PRIMARY KEY,
    author_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_created ON chat_messages(created_at DESC);
