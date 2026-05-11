-- Восстанавливаем rss_articles из внешних постов и откатываем изменения posts.

CREATE TABLE rss_articles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id       UUID NOT NULL REFERENCES rss_sources(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    url             TEXT NOT NULL UNIQUE,
    summary         TEXT NOT NULL DEFAULT '',
    cover_image_url TEXT NOT NULL DEFAULT '',
    tags            TEXT[] NOT NULL DEFAULT '{}',
    published_at    TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX rss_articles_published_at_idx ON rss_articles (published_at DESC);
CREATE INDEX rss_articles_tags_idx         ON rss_articles USING GIN (tags);
CREATE INDEX rss_articles_source_id_idx    ON rss_articles (source_id);

INSERT INTO rss_articles (id, source_id, title, url, summary, cover_image_url, tags, published_at, created_at)
SELECT id, source_id, title, external_url, content, cover_image_url, tags, created_at, created_at
FROM posts WHERE source_id IS NOT NULL;

DELETE FROM posts WHERE source_id IS NOT NULL;

ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_origin_check;
ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_external_url_unique;
DROP INDEX IF EXISTS idx_posts_tags;
DROP INDEX IF EXISTS idx_posts_external_at;
DROP INDEX IF EXISTS idx_posts_source;
ALTER TABLE posts DROP COLUMN tags;
ALTER TABLE posts DROP COLUMN cover_image_url;
ALTER TABLE posts DROP COLUMN source_id;
ALTER TABLE posts DROP COLUMN external_url;
ALTER TABLE posts ALTER COLUMN author_id    SET NOT NULL;
ALTER TABLE posts ALTER COLUMN community_id SET NOT NULL;
