-- community rules (jsonb array of strings)
ALTER TABLE communities ADD COLUMN rules JSONB NOT NULL DEFAULT '[]';

CREATE TABLE community_moderators (
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         TEXT NOT NULL DEFAULT 'moderator' CHECK (role IN ('moderator', 'admin')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (community_id, user_id)
);

CREATE INDEX idx_community_mods_user ON community_moderators(user_id);
