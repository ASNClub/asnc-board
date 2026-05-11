-- honeydrop initial schema
-- golang-migrate: UP

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth_id      TEXT NOT NULL UNIQUE,
    username     TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    avatar_url   TEXT,
    banner_url   TEXT,
    bio          TEXT,
    reputation   INT  NOT NULL DEFAULT 0,
    privacy      TEXT NOT NULL DEFAULT 'public' CHECK (privacy IN ('public', 'private')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_tags (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag     TEXT NOT NULL,
    PRIMARY KEY (user_id, tag)
);

-- привязанные соцсети (для отображения на профиле)
CREATE TABLE user_platforms (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    username    TEXT,
    profile_url TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_follows (
    follower_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id != following_id)
);

CREATE TABLE friendships (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (requester_id, addressee_id),
    CHECK (requester_id != addressee_id)
);

-- подключённые git-хостинги (OAuth токен или PAT)
CREATE TABLE user_git_accounts (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider     TEXT NOT NULL, -- forgejo | gitea | github
    access_token TEXT NOT NULL, -- зашифрован на уровне приложения
    username     TEXT NOT NULL,
    instance_url TEXT,          -- для self-hosted: https://git.example.com
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, provider, instance_url)
);

-- закреплённые репозитории на профиле
CREATE TABLE pinned_repos (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    git_account_id UUID NOT NULL REFERENCES user_git_accounts(id) ON DELETE CASCADE,
    external_id    TEXT NOT NULL,    -- ID репо на хостинге
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

CREATE TABLE communities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    slug            TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    description     TEXT,
    avatar_url      TEXT,
    banner_url      TEXT,
    followers_count INT  NOT NULL DEFAULT 0,
    posts_count     INT  NOT NULL DEFAULT 0,
    stars_count     INT  NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id) -- одно комьюнити на пользователя
);

CREATE TABLE community_tags (
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    tag          TEXT NOT NULL,
    PRIMARY KEY (community_id, tag)
);

CREATE TABLE community_follows (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, community_id)
);

-- звёзды (как GitHub stars, только подписчики)
CREATE TABLE community_stars (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, community_id)
);

-- бан / мут пользователя в комьюнити
CREATE TABLE community_bans (
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type         TEXT NOT NULL CHECK (type IN ('ban', 'mute')),
    reason       TEXT,
    expires_at   TIMESTAMPTZ,  -- NULL = перманентно
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (community_id, user_id)
);

CREATE TABLE posts (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    author_id    UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title        TEXT,
    content      TEXT NOT NULL, -- Markdown
    views_count  INT  NOT NULL DEFAULT 0,
    votes_count  INT  NOT NULL DEFAULT 0,
    is_pinned    BOOL NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE post_media (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    type       TEXT NOT NULL CHECK (type IN ('image', 'video', 'file')),
    url        TEXT NOT NULL,  -- путь в S3
    name       TEXT NOT NULL,  -- оригинальное имя файла
    size       INT  NOT NULL,  -- байты
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE post_votes (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE comments (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id     UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    parent_id   UUID REFERENCES comments(id) ON DELETE CASCADE,
    content     TEXT NOT NULL, -- Markdown
    votes_count INT  NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE comment_votes (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, comment_id)
);

-- Github-like, заготовка на будущее
CREATE TABLE sponsorships (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sponsor_id      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    target_type     TEXT NOT NULL CHECK (target_type IN ('user', 'community')),
    target_id       UUID NOT NULL, -- users.id или communities.id
    amount          INT  NOT NULL, -- в копейках
    currency        TEXT NOT NULL DEFAULT 'RUB',
    provider        TEXT NOT NULL, -- yukassa | tinkoff | ...
    provider_sub_id TEXT,          -- ID рекуррентной подписки на стороне провайдера
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'cancelled', 'expired')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    idempotency_key TEXT NOT NULL UNIQUE,
    sponsorship_id  UUID NOT NULL REFERENCES sponsorships(id) ON DELETE RESTRICT,
    provider        TEXT NOT NULL,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    payload    JSONB NOT NULL DEFAULT '{}',
    is_read    BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notification_settings (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type    TEXT NOT NULL,
    enabled BOOL NOT NULL DEFAULT TRUE,
    PRIMARY KEY (user_id, type)
);

CREATE TABLE activity_events (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    -- post.created | comment.created | community.followed
    -- community.starred | user.followed | friendship.accepted
    payload    JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_username         ON users(username);
CREATE INDEX idx_users_auth_id          ON users(auth_id);

CREATE INDEX idx_user_follows_following ON user_follows(following_id);
CREATE INDEX idx_friendships_addressee  ON friendships(addressee_id);
CREATE INDEX idx_friendships_requester  ON friendships(requester_id);

CREATE INDEX idx_user_git_accounts_user ON user_git_accounts(user_id);
CREATE INDEX idx_pinned_repos_user      ON pinned_repos(user_id, sort_order);

CREATE INDEX idx_communities_slug       ON communities(slug);
CREATE INDEX idx_communities_owner      ON communities(owner_id);
CREATE INDEX idx_community_follows_comm ON community_follows(community_id);
CREATE INDEX idx_community_stars_comm   ON community_stars(community_id);

CREATE INDEX idx_posts_community        ON posts(community_id, created_at DESC);
CREATE INDEX idx_posts_author           ON posts(author_id, created_at DESC);
CREATE INDEX idx_post_votes_post        ON post_votes(post_id);

CREATE INDEX idx_comments_post          ON comments(post_id, created_at);
CREATE INDEX idx_comments_parent        ON comments(parent_id);
CREATE INDEX idx_comment_votes_comment  ON comment_votes(comment_id);

CREATE INDEX idx_sponsorships_sponsor   ON sponsorships(sponsor_id);
CREATE INDEX idx_sponsorships_target    ON sponsorships(target_type, target_id);

CREATE INDEX idx_notifications_user     ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_unread   ON notifications(user_id, is_read) WHERE is_read = FALSE;

CREATE INDEX idx_activity_user          ON activity_events(user_id, created_at DESC);
