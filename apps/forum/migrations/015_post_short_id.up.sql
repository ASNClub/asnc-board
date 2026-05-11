-- Add short_id to posts for shareable URLs (/p/<shortId>).
-- 8-char hex from md5(random + id) — guaranteed different per row, no extension needed.

ALTER TABLE posts ADD COLUMN IF NOT EXISTS short_id VARCHAR(12);

UPDATE posts
SET short_id = substr(md5(random()::text || id::text || clock_timestamp()::text), 1, 8)
WHERE short_id IS NULL;

ALTER TABLE posts ALTER COLUMN short_id SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_posts_short_id ON posts(short_id);
