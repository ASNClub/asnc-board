-- GitHub / GitLab / Codeberg OAuth-интеграции.
-- Пересоздаём таблицы, снесённые в 008, с полями под OAuth (refresh_token, expires_at).
CREATE TABLE user_git_accounts (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider      TEXT NOT NULL,                  -- 'github' | 'gitlab' | 'codeberg'
    access_token  TEXT NOT NULL,
    refresh_token TEXT,
    expires_at    TIMESTAMPTZ,                    -- NULL = токен бессрочный (github classic / codeberg)
    username      TEXT NOT NULL,
    instance_url  TEXT,                           -- для self-hosted gitlab
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, provider, instance_url)
);

CREATE TABLE pinned_repos (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    git_account_id UUID NOT NULL REFERENCES user_git_accounts(id) ON DELETE CASCADE,
    external_id    TEXT NOT NULL,
    name           TEXT NOT NULL,
    description    TEXT,
    url            TEXT NOT NULL,
    language       TEXT,
    stars_count    INT  NOT NULL DEFAULT 0,
    forks_count    INT  NOT NULL DEFAULT 0,
    is_fork        BOOL NOT NULL DEFAULT FALSE,
    topics         TEXT[] NOT NULL DEFAULT '{}',
    sort_order     INT  NOT NULL DEFAULT 0,
    synced_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, git_account_id, external_id)
);

CREATE INDEX idx_user_git_accounts_user ON user_git_accounts(user_id);
CREATE INDEX idx_pinned_repos_user      ON pinned_repos(user_id, sort_order);
