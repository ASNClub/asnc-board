-- honeydrop feed schema
-- golang-migrate: UP

CREATE TABLE rss_sources (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    url          TEXT NOT NULL UNIQUE,
    name         TEXT NOT NULL,
    site_url     TEXT NOT NULL DEFAULT '',
    favicon_url  TEXT NOT NULL DEFAULT '',
    tags         TEXT[] NOT NULL DEFAULT '{}',
    last_fetched TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX rss_sources_tags_idx ON rss_sources USING GIN (tags);

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
CREATE INDEX rss_articles_tags_idx ON rss_articles USING GIN (tags);
CREATE INDEX rss_articles_source_id_idx ON rss_articles (source_id);
