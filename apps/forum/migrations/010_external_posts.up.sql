ALTER TABLE posts ALTER COLUMN community_id DROP NOT NULL;
ALTER TABLE posts ALTER COLUMN author_id    DROP NOT NULL;
ALTER TABLE posts ADD COLUMN external_url    TEXT;
ALTER TABLE posts ADD COLUMN source_id       UUID REFERENCES rss_sources(id) ON DELETE SET NULL;
ALTER TABLE posts ADD COLUMN cover_image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE posts ADD COLUMN tags            TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE posts ADD CONSTRAINT posts_external_url_unique UNIQUE (external_url);
CREATE INDEX idx_posts_source       ON posts(source_id, created_at DESC) WHERE source_id IS NOT NULL;
CREATE INDEX idx_posts_external_at  ON posts(created_at DESC)            WHERE source_id IS NOT NULL;
CREATE INDEX idx_posts_tags         ON posts USING GIN (tags);

-- Внешние посты не могут быть закреплены (комьюнити нет), но constraint оставим только на форум-посты.
ALTER TABLE posts ADD CONSTRAINT posts_origin_check CHECK (
    (source_id IS NOT NULL AND author_id IS NULL  AND community_id IS NULL  AND external_url IS NOT NULL)
    OR
    (source_id IS NULL     AND author_id IS NOT NULL AND community_id IS NOT NULL)
);

-- Перелив существующих rss_articles в posts.
INSERT INTO posts (id, community_id, author_id, title, content, votes_count,
                   source_id, external_url, cover_image_url, tags, created_at, updated_at)
SELECT
    ra.id,
    NULL,
    NULL,
    ra.title,
    ra.summary,
    0,
    ra.source_id,
    ra.url,
    ra.cover_image_url,
    ra.tags,
    ra.published_at,
    ra.published_at
FROM rss_articles ra
ON CONFLICT (id) DO NOTHING;

DROP TABLE rss_articles;
